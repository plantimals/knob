# knob

![knob logo](https://plantimals.org/img/knob-nostr.png)

named after [Knob Noster](https://en.wikipedia.org/wiki/Knob_Noster%2C_Missouri), knob is a command line tool for
generating new priv/pub key pairs and publishing new `kind 1` messages to 
nostr. there are several options for input, from the `--input` command line
flag to `.json`, `.md`, or `.txt` files.

check out the repo, then grab dependencies:

```shell
go mod download
```

then build:

```shell
make 
```

post a message from the command line directly with an existing private key
(specified by NOSTR_KEY env variable):

```shell
NOSTR_KEY=fc6258cf0456c6ad658eab9f329a2fe7dda271ac08352e4256fd73f0b75c4dbf ./knob --input "this is my message"
```

post a message from the command line with newly generated keys:

```shell
./knob --genkeys --input "this is my message"
```

post a message from event(s) in a `.json` file. specify your own relay (default is `nostr.drss.io`):

```shell
NOSTR_KEY=fc6258cf0456c6ad658eab9f329a2fe7dda271ac08352e4256fd73f0b75c4dbf ./knob --file events.json --relay "wss://relay.damus.io"
```
