# Python Tag Calculator

A simple angelina client that calculates tag combinations in python3.
The code is not very robust and is not recommended for normal users, though it does work fine.

## Usage

Run `main.py` after angelina is started and the game data finishes updating.
The client will connect with angelina on startup and automatically attach onto the latest connected user.
Tag calculation will be run on all free recruitment slots whenever the user enters the recruitment page, finishes a recruitment, or refreshes tags in a recruitment slot.
The client prints the tags in any empty recruitment slots, followed by the combinations that guarantee a 4* (or above) character.

```sh
$ python main.py
loading tag data...
loading character data...
ready
attached to GL_14760000
slot #1: ['Sniper', 'Ranged', 'Healing', 'Starter', 'DPS']:
slot #1: ['Defender', 'Supporter', 'Caster', 'Ranged', 'Defense']:
slot #1: ['Defender', 'Medic', 'Supporter', 'Caster', 'Slow']:
slot #1: ['Caster', 'Specialist', 'Melee', 'Starter', 'Debuff']:
['Specialist'] -> ['FEater', 'Shaw', 'Rope', 'Gravel', 'Cliffheart', 'Manticore', 'Projekt Red']
['Debuff'] -> ['Pramanix', 'Meteorite', 'Ifrit', 'Haze', 'Meteor']
['Caster', 'Debuff'] -> ['Ifrit', 'Haze']
['Specialist', 'Melee'] -> ['FEater', 'Shaw', 'Rope', 'Gravel', 'Cliffheart', 'Manticore', 'Projekt Red']
slot #1: ['Defender', 'Caster', 'Melee', 'Healing', 'Fast-Redeploy']:
['Fast-Redeploy'] -> ['Gravel', 'Projekt Red']
['Defender', 'Healing'] -> ['Saria', 'Nearl', 'Gummy']
['Melee', 'Healing'] -> ['Saria', 'Nearl', 'Gummy']
['Melee', 'Fast-Redeploy'] -> ['Gravel', 'Projekt Red']
['Defender', 'Melee', 'Healing'] -> ['Saria', 'Nearl', 'Gummy']
slot #1: ['Sniper', 'Supporter', 'Caster', 'DP-Recovery', 'Defense']:
```

## Note to developers

People who might want to improve on the client's design should be aware of the following caveats in this demo:
- The state handling is extremely naive due to the simplistic nature of the module, since it just needs to handle recruitment data. A very simple dispatch system is implemented, e.g., responses from a C_Get request are only ever routed to `Recruit.parse_tags()`.
- It prints tag data from every open slot, regardless if it has already printed it before, e.g., refreshing a tag with 2 slots open will print the tags from both slots.
- No timeouts/retries are implemented in the demo. Getting the game data statically served by angelina might fail if it's not been updated yet. Websockets are reliable but the initial connection may fail if angelina is not ready.
- Developers could look into [asyncio.Event](https://docs.python.org/3/library/asyncio-sync.html#asyncio.Event) and event emitting packages such as [pyee](https://github.com/jfhbrook/pyee) for better handling of events.
