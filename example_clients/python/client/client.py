# coding=utf-8

import asyncio
import gzip
import json
import re
import time
from itertools import combinations
from typing import Callable, Dict, List, Set, Tuple
from urllib.request import Request, urlopen

from websockets import WebSocketClientProtocol
from websockets.exceptions import ConnectionClosed, ConnectionClosedError

GACHA_TABLE_URL = "https://raw.githubusercontent.com/Kengxxiao/ArknightsGameData/master/en_US/gamedata/excel/gacha_table.json"
CHAR_TABLE_URL = "https://raw.githubusercontent.com/Kengxxiao/ArknightsGameData/master/en_US/gamedata/excel/character_table.json"

PROFESSION_TO_TAG_ID = {
    "MEDIC": 4,
    "WARRIOR": 1,
    "PIONEER": 8,
    "TANK": 3,
    "SNIPER": 2,
    "CASTER": 6,
    "SUPPORT": 5,
    "SPECIAL": 7,
}
POSITION_TO_TAG_ID = {
    "MELEE": 9,
    "RANGED": 10,
}
RARITY_TO_TAG_ID = {
    5: 11,
    4: 14,
}
ROBOT_TAG_ID = 28
# Only print tag combinations at this rarity or above,
# where a rarity of 3 corresponds to 4*
RARITY_THRESH = 3
STAR_TOK = "★"


class Client:
    def __init__(self, ws: WebSocketClientProtocol, loop: asyncio.AbstractEventLoop):
        self.ws = ws
        self.loop = loop
        self.hooks: List[str] = []
        self.handlers: Dict[str, Callable[[str], None]] = {
            "S_UserList": self.handle_userlist,
            "S_NewUser": self.handle_newuser,
            "S_Attached": self.handle_attached,
            "S_Detached": self.handle_detach,
            "S_HookEvt": self.handle_hookevt,
            "S_Get": self.handle_get,
            "S_Hooked": self.handle_dummy,
            "S_Error": self.handle_dummy,
        }
        self.users: Set[str] = set()
        self.attached_user = ""
        # Load data required for parsing tags
        self._init_recruit_data()

    def _init_recruit_data(self):
        print("loading tag data...")
        data = json.loads(get_url(GACHA_TABLE_URL))
        # parse_recruitable_chars is incredibly janky and can break at any time
        # or perhaps for different languages. But it's a necessity.
        self.recruitableCharNames = parse_recruitable_chars(data["recruitDetail"])
        self.tagIdToName = {v["tagId"]: v["tagName"] for v in data["gachaTags"]}
        tagNameToId = {v: k for k, v in self.tagIdToName.items()}
        tagIdToOpSet = {k: set() for k in self.tagIdToName}

        print("loading character data...")
        data = json.loads(get_url(CHAR_TABLE_URL))
        char_data = {}
        for k, v in data.items():
            # Some UGLY additional data processing
            if v["tagList"] is None:
                continue
            name = v["name"]
            # Case insensitive check because of inconsistency in the client,
            # e.g. FEater -> Feater.
            if name.lower() not in self.recruitableCharNames:
                continue
            data = {
                "name": v["name"],
                "rarity": v["rarity"],
            }
            tags = [tagNameToId[tag_name] for tag_name in v["tagList"]]
            tags.append(PROFESSION_TO_TAG_ID[v["profession"]])
            tags.append(POSITION_TO_TAG_ID[v["position"]])
            if v["displayNumber"].startswith("RCX"):
                # Hardcoding for robots
                tags.append(ROBOT_TAG_ID)
            if v["rarity"] in RARITY_TO_TAG_ID:
                tags.append(RARITY_TO_TAG_ID[v["rarity"]])
            data["tags"] = tags
            char_data[k] = data
        self.char_data = char_data

        for char_id, char_data in self.char_data.items():
            for tag_id in char_data["tags"]:
                tagIdToOpSet[tag_id].add(char_id)

        self.tagNameToId = tagNameToId
        self.tagIdToOpSet = tagIdToOpSet
        print("done")

    def print_results(self, comb: Tuple[int], id_set: Set):
        # Print results from analyzing recruitment tags.
        # Params:
        #   - comb : a tupple of tag IDs
        #   - id_set: a set of character IDs that match the combination
        filtered_chars: List[str] = []
        min_rarity = 10
        high_rarity = -1
        for char_id in id_set:
            char = self.char_data[char_id]
            rarity = char["rarity"]
            if rarity < min_rarity:
                min_rarity = rarity
            if rarity > high_rarity:
                high_rarity = rarity
            if rarity == 0:
                # Robot!
                filtered_chars.append(char["name"])
            elif rarity >= RARITY_THRESH:
                filtered_chars.append(char["name"])

        if (min_rarity == 0 and high_rarity != 0) or min_rarity < RARITY_THRESH:
            # Return if it's not a robot-only tag or the min rarity is below our thresh
            return

        # Only print the tag combinations if we find something good!
        if len(filtered_chars) > 0:
            tag_names = [self.tagIdToName[v] for v in comb]
            print(f"{tag_names} -> {filtered_chars}")

    async def handle_userlist(self, payload: str):
        users: List[str] = json.loads(payload)
        if len(users) > 0:
            self.users = set(users)
            # Attempt to attach to the most recently connected user.
            await self.send_attach(users[-1])
        # Remove the handler
        return self.handlers["S_UserList"]

    # Dummy handler for server packets we're not really interested in.
    async def handle_dummy(self, payload: str):
        pass

    async def handle_newuser(self, payload: str):
        user: str = json.loads(payload)
        self.users.add(user)
        # Attempt to attach to a any new user
        if self.attached_user != "":
            await self.send_detach()
        await self.send_attach(user)

    async def handle_get(self, payload: str):
        start_t = time.time()
        slots: Dict[str, Dict] = json.loads(payload)["data"]
        for slot_n, slot in slots.items():
            if slot["state"] != 1:
                # State 0 and 2 are for when the slot is locked and busy respectively.
                continue
            tag_choices = slot["tags"]
            print(f"slot #{slot_n}: {[self.tagIdToName[v] for v in tag_choices]}:")
            n_choices = len(tag_choices)
            # Consider combinations for tags (1,4)
            for nTags in range(1, 4):
                for tagComb in combinations(range(n_choices), nTags):
                    tagComb = tuple(tag_choices[v] for v in tagComb)
                    char_sets: List[Set] = []
                    for tag in tagComb:
                        char_sets.append(self.tagIdToOpSet[tag])
                    if len(char_sets) == 1:
                        self.print_results(tagComb, char_sets[0])
                        continue
                    # If the intersection of sets is not trivial then...
                    result = char_sets[0].intersection(*char_sets[1:])
                    if len(result) > 0:
                        self.print_results(tagComb, result)

    async def handle_attached(self, payload: str):
        # Successfully attached to a user.
        user = json.loads(payload)
        self.attached_user = user
        # We'll now initialize the hooks
        print(f"attached to {user}")
        await self.send_hook("packet", "S/gacha/syncNormalGacha", True)
        await self.send_hook("packet", "S/gacha/finishNormalGacha", True)
        await self.send_hook("packet", "S/gacha/refreshTags", True)

    async def handle_detach(self, payload: str):
        self.attached_user = ""

    async def handle_hookevt(self, payload: str):
        # We initialized the 2 hooks earlier as event hooks, so they won't receive
        # any useful information. We'll just know that the client has received
        # either syncNormalGacha (sent when entering the recruitment page) or
        # finishNormalGacha (finishing a recruitment).
        await self.send_get("recruit.normal.slots")

    async def send_get(self, target: str):
        await self.ws.send("C_Get " + json.dumps(target))

    async def send_attach(self, user: str):
        # Attempt to attach ourselves to a user.
        await self.ws.send(f"C_Attach {json.dumps(user)}")

    async def send_hook(self, type: str, target: str, event: bool):
        await self.ws.send(
            "C_Hook " + json.dumps({"type": type, "target": target, "event": event})
        )

    async def send_detach(self):
        await self.ws.send("C_Detach")

    async def recv_loop(self):
        try:
            while True:
                s: str = await self.ws.recv()
                toks = s.split(" ", maxsplit=1)
                op = toks[0]
                payload = ""
                n_toks = len(toks)
                if n_toks >= 2:
                    payload = toks[1]
                handler = self.handlers.get(op)
                if handler is None:
                    print(f"handler not found for {op}")
                    continue
                await handler(payload)  # type:ignore
        except (ConnectionClosed, ConnectionClosedError):
            pass

    def shutdown(self):
        tasks = self.loop.create_task(self.ws.close())
        self.loop.run_until_complete(asyncio.gather(tasks))


def get_url(url: str):
    req = Request(url)
    req.add_header("Accept-Encoding", "gzip")
    response = urlopen(req)
    content = gzip.decompress(response.read())
    return content.decode("utf-8")


def parse_recruitable_chars(s: str) -> Set[str]:
    # Returns the names of all the units in the recruitment pool in lower case.
    # Parsing relies on the star token delimiting rarities and that they will
    # always have a divider with "-" characters between the rarities.
    ret = set()
    min_pos = s.find(STAR_TOK + "\n")
    for rarity in range(1, 7):
        start_s = STAR_TOK * rarity + "\n"
        start_pos = s.find(start_s, min_pos) + len(start_s)
        end_pos = s.find("\n-", start_pos)
        s2 = s[start_pos:end_pos]
        min_pos = end_pos
        # Remove unity markup
        s2 = re.sub(r"<.*?>", "", s2)
        sl = s2.split("/")
        for v in sl:
            ret.add(v.strip().lower())
    return ret
