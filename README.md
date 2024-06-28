# tf2bd_cleaner

Removes VAC, game banned, and deleted profiles from [tf2bd](https://github.com/PazerOP/tf2_bot_detector) lists.

Can also summarize basic stats from the playerlist without making any changes.

## Examples

Get stats on the list without making any changes:
```shell
./tf2bd_cleaner -k your-api-key < testdata/playerlist.cleffy.json -s
2024/06/27 23:16:53 INFO Running profile checks...
2024/06/27 23:16:59 INFO Profile stats total=909 game_bans=12 vac_bans=94 community_bans=131 deleted=247 valid=425
2024/06/27 23:16:59 INFO Tag stats suspicious=12 cheater=897

```

Include community bans in removal (-c). These are not removed by default as they can be a non-permanent account status:

```shell
./tf2bd_cleaner -k your-api-key -i testdata/playerlist.cleffy.json -o playerlist.trimmed-cleffy.json -c
```

Write results to a new file:

```shell
./tf2bd_cleaner -k your-api-key -i testdata/playerlist.cleffy.json -o playerlist.trimmed-cleffy.json
```

Overwrite/replace the input file with the new player list. This is destructive, so make sure you backup your files if you want to be safe:

```shell
./tf2bd_cleaner -k your-api-key -i testdata/playerlist.cleffy.json -r
```

Process data using stdin/stdout:

```shell
./tf2bd_cleaner -k your-api-key < playerlist.cleffy.json > playerlist.out.json
``` 

Full Example:

    $ ./tf2bd_cleaner -k xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx -i testdata/playerlist.cleffy.json -o playerlist.trimmed.json -c 
    2024/06/27 23:20:37 INFO Running profile checks...
    2024/06/27 23:20:41 INFO Profile stats total=909 game_bans=12 vac_bans=94 community_bans=131 deleted=247 valid=425
    2024/06/27 23:20:41 INFO Tag stats suspicious=12 cheater=897
    $


## Usage Help

    Remove banned and deleted users from TF2 bot detector player lists
    
    Usage:
    tf2bd_cleaner [flags]
    
    Flags:
    -k, --apikey string   Steam API Key
    -c, --community       Include community bans in deletions
    -h, --help            help for tf2bd_cleaner
    -i, --input string    Input player list path. If not defined, stdin will be used
    -o, --output string   Output player list path. If not defined, stdout will be used
    -r, --overwrite       Overwrite the input file
    -s, --stats           Computes stats for entries, does not perform any deletions
