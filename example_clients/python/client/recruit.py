# coding=utf-8

from gzip import decompress
from itertools import combinations
from json import loads
from re import sub
from typing import Dict, List, Set, Tuple
from urllib.request import Request, urlopen

GACHA_TABLE_URL = "https://raw.githubusercontent.com/Kengxxiao/ArknightsGameData/master/en_US/gamedata/excel/gacha_table.json"
CHAR_TABLE_URL = "https://raw.githubusercontent.com/Kengxxiao/ArknightsGameData/master/en_US/gamedata/excel/character_table.json"

PROFESSION_TO_TAG = {
    "MEDIC": 4,
    "WARRIOR": 1,
    "PIONEER": 8,
    "TANK": 3,
    "SNIPER": 2,
    "CASTER": 6,
    "SUPPORT": 5,
    "SPECIAL": 7,
}
POSITION_TO_TAG = {
    "MELEE": 9,
    "RANGED": 10,
}
RARITY_TO_TAG = {
    5: 11,
    4: 14,
}
ROBOT_TAG = 28
# Only print tag combinations at this rarity or above,
# where a rarity of 3 corresponds to 4*
RARITY_THRESH = 3
# Star character used for parsing recruitable characters.
STAR_TOK = "★"


class Recruit:
    """
    The Recruit class retrieves and parses all the game data from online sources
    required to parse the tags in a relatively client region agnostic way. It should
    still work if you swap the GACHA_TABLE_URL and CHAR_TABLE_URL constants to a new
    region.

    Once initialized, the caller can call parse_tags with the payload of a C_Get
    whose target is 'recruit.normal.slots' to print tag combinations.
    """

    def __init__(self):
        print("loading tag data...")
        data = loads(get_url(GACHA_TABLE_URL))
        # parse_recruitable_chars is incredibly janky and can break at any time
        # or perhaps for different languages. But it's a necessity.
        self.recruitable = parse_recruitable_chars(data["recruitDetail"])
        self.tag_to_name = {v["tagId"]: v["tagName"] for v in data["gachaTags"]}
        name_to_tag = {v: k for k, v in self.tag_to_name.items()}
        tagIdToOpSet = {k: set() for k in self.tag_to_name}

        print("loading character data...")
        data = loads(get_url(CHAR_TABLE_URL))
        char_data = {}
        for k, v in data.items():
            # Some UGLY data processing
            if v["tagList"] is None:
                continue
            name = v["name"]
            # Case insensitive check because of inconsistency in the GL client,
            # e.g. FEater -> Feater.
            if name.lower() not in self.recruitable:
                continue
            data = {
                "name": v["name"],
                "rarity": v["rarity"],
            }
            tags = [name_to_tag[tag_name] for tag_name in v["tagList"]]
            tags.append(PROFESSION_TO_TAG[v["profession"]])
            tags.append(POSITION_TO_TAG[v["position"]])
            if v["displayNumber"].startswith("RCX"):
                # Hardcoding for robots
                tags.append(ROBOT_TAG)
            if v["rarity"] in RARITY_TO_TAG:
                tags.append(RARITY_TO_TAG[v["rarity"]])
            data["tags"] = tags
            char_data[k] = data
        self.char_data = char_data

        for char_id, char_data in self.char_data.items():
            for tag_id in char_data["tags"]:
                tagIdToOpSet[tag_id].add(char_id)

        self.tagNameToId = name_to_tag
        self.tagIdToOpSet = tagIdToOpSet
        print("done")

    def parse_tags(self, payload: str):
        slots: Dict[str, Dict] = loads(payload)["data"]
        for slot_n, slot in slots.items():
            if slot["state"] != 1:
                # State 0 and 2 are for when the slot is locked and busy respectively.
                continue
            tag_choices = slot["tags"]
            print(f"slot #{slot_n}: {[self.tag_to_name[v] for v in tag_choices]}:")
            n_choices = len(tag_choices)
            # Consider combinations for tags (1,4)
            for nTags in range(1, 4):
                for tagComb in combinations(range(n_choices), nTags):
                    tagComb = tuple(tag_choices[v] for v in tagComb)
                    char_sets: List[Set] = []
                    for tag in tagComb:
                        char_sets.append(self.tagIdToOpSet[tag])
                    if len(char_sets) == 1:
                        self._print_results(tagComb, char_sets[0])
                        continue
                    # If the intersection of sets is not trivial then...
                    result = char_sets[0].intersection(*char_sets[1:])
                    if len(result) > 0:
                        self._print_results(tagComb, result)

    def _print_results(self, comb: Tuple[int], id_set: Set):
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
            tag_names = [self.tag_to_name[v] for v in comb]
            print(f"{tag_names} -> {filtered_chars}")


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
        s2 = sub(r"<.*?>", "", s2)
        sl = s2.split("/")
        for v in sl:
            ret.add(v.strip().lower())
    return ret


def get_url(url: str):
    req = Request(url)
    req.add_header("Accept-Encoding", "gzip")
    response = urlopen(req)
    content = decompress(response.read())
    return content.decode("utf-8")
