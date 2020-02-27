# coding=utf-8

import asyncio
from json import dumps, loads
from typing import Awaitable, Callable, Dict, List

import websockets
from websockets.exceptions import ConnectionClosed, ConnectionClosedError

from .recruit import Recruit


class Client:
    def __init__(self, base_url: str):
        self.loop = asyncio.get_event_loop()
        self.handlers: Dict[str, Callable[[str, str], Awaitable[None]]] = {
            "S_UserList": self.handle_userlist,
            "S_NewUser": self.handle_newuser,
            "S_Attached": self.handle_attached,
            "S_Detached": self.handle_detach,
            "S_HookEvt": self.handle_hookevt,
            "S_Get": self.handle_get,
            "S_Hooked": self.handle_dummy,
        }
        self.attached_user = ""
        self.base_url = base_url
        # Load data required for parsing tags
        self.recruit = Recruit(base_url)

    async def handle_userlist(self, op: str, payload: str):
        users: List[str] = loads(payload)
        if len(users) > 0:
            # Attempt to attach to the most recently connected user.
            await self.send_attach(users[-1])
        # Remove the handler
        return self.handlers["S_UserList"]

    # Dummy handler for server packets we're not really interested in.
    async def handle_dummy(self, op: str, payload: str):
        pass

    async def handle_newuser(self, op: str, payload: str):
        user: str = loads(payload)
        # Attempt to attach to a new user
        if self.attached_user != "":
            await self.send_detach()
        await self.send_attach(user)

    async def handle_get(self, op: str, payload: str):
        self.recruit.parse_tags(payload)

    async def handle_attached(self, op: str, payload: str):
        # Successfully attached to a user.
        user = loads(payload)
        self.attached_user = user
        print(f"attached to {user}")
        # Initialize the hooks to notify us when the user enters the recruitment page,
        # finishes a recruitment, or refreshes the tags.
        await self.send_hook("packet", "S/gacha/syncNormalGacha", True)
        await self.send_hook("packet", "S/gacha/finishNormalGacha", True)
        await self.send_hook("packet", "S/gacha/refreshTags", True)

    async def handle_detach(self, op: str, payload: str):
        self.attached_user = ""

    async def handle_hookevt(self, op: str, payload: str):
        # We initialized the 2 hooks earlier as event hooks, so they won't receive
        # any useful information. We'll just know that the client has received
        # either syncNormalGacha (sent when entering the recruitment page) or
        # finishNormalGacha (finishing a recruitment).
        await self.send_get("recruit.normal.slots")

    async def send_get(self, target: str):
        await self.ws.send("C_Get " + dumps(target))

    async def send_attach(self, user: str):
        # Attempt to attach ourselves to a user.
        await self.ws.send(f"C_Attach {dumps(user)}")

    async def send_hook(self, type: str, target: str, event: bool):
        await self.ws.send(
            "C_Hook " + dumps({"type": type, "target": target, "event": event})
        )

    async def send_detach(self):
        await self.ws.send("C_Detach")

    async def run(self):
        async with websockets.connect(f"ws://{self.base_url}/ws") as ws:
            self.ws = ws
            try:
                while True:
                    s: str = await ws.recv()
                    toks = s.split(" ", maxsplit=1)
                    op = toks[0]
                    n_toks = len(toks)
                    if n_toks >= 2:
                        payload = toks[1]
                    else:
                        payload = ""
                    handler = self.handlers.get(op)
                    if handler is None:
                        print(f"Unhandled op:{op}, payload:{payload}")
                        continue
                    await handler(op, payload)
            except (ConnectionClosed, ConnectionClosedError):
                pass

    def shutdown(self):
        tasks = self.loop.create_task(self.ws.close())
        self.loop.run_until_complete(asyncio.gather(tasks))
