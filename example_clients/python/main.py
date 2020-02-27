#!/usr/bin/env python3
# coding=utf-8

import asyncio

from client import Client

client = Client("localhost:8000")


try:
    asyncio.get_event_loop().run_until_complete(client.run())
except KeyboardInterrupt:
    client.shutdown()
