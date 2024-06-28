# tf2bd_cleaner

Removes VAC, game banned, and deleted profiles from [tf2bd](https://github.com/PazerOP/tf2_bot_detector) lists.

## Usage

Write results to a new file:

```shell
./tf2bd_cleaner -k your-api-key -i testdata/playerlist.cleffy.json -o playerlist.trimmed-cleffy.json
```

Overwrite/replace the input file with the new player list. This is destructive, so make sure you backup your files if you want to be safe:

```shell
./tf2bd_cleaner -k your-api-key -i testdata/playerlist.cleffy.json -r
```

Process from standard input:

```shell
./tf2bd_cleaner -k your-api-key < playerlist.cleffy.json > playerlist.out.json
``` 

Help overview:

    Remove banned and deleted users from TF2 bot detector player lists

    Usage:
    tf2bd_cleaner [flags]
    
    Flags:
    -k, --apikey string   Steam API Key
    -h, --help            help for tf2bd_cleaner
    -i, --input string    Input player list path. If not defined, stdin will be used
    -o, --output string   Output player list path. If not defined, stdout will be used
    -r, --overwrite       Overwrite the input file.
