#!/usr/bin/env python
# coding=utf-8

import asyncio
import websockets
from client import Client
loop = asyncio.get_event_loop()

client = Client(None, loop)

async def main(uri: str):
    loop = asyncio.get_event_loop()
    async with websockets.connect(uri) as websocket:
        client.ws = websocket
        recv_loop = loop.create_task(client.recv_loop())
        await asyncio.gather(recv_loop)

try:
    asyncio.get_event_loop().run_until_complete(main('ws://localhost:8000/ws'))
except KeyboardInterrupt:
    client.shutdown()
