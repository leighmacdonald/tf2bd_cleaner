// This tool accepts a TF2 Bot Detector compatible json schema as input. It checks marked players
// against the Steam WebAPI and will remove players that are either marked as VAC banned or outright deleted.

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"slices"

	"github.com/leighmacdonald/steamid/v4/steamid"
	"github.com/leighmacdonald/steamweb/v2"
	"github.com/spf13/cobra"
)

type FileInfo struct {
	Authors     []string `json:"authors"`
	Description string   `json:"description"`
	Title       string   `json:"title"`
	UpdateURL   string   `json:"update_url"`
}

type LastSeen struct {
	PlayerName string `json:"player_name,omitempty"`
	Time       int    `json:"time,omitempty"`
}

type Players struct {
	Attributes []string `json:"attributes"`
	LastSeen   LastSeen `json:"last_seen,omitempty"`
	Steamid    any      `json:"steamid"`
	Proof      []string `json:"proof,omitempty"`
}

type TF2BDSchema struct {
	Schema   string    `json:"$schema"` //nolint:tagliatelle
	FileInfo FileInfo  `json:"file_info"`
	Players  []Players `json:"players"`
}

var (
	apiKey     string
	inputFile  string
	outputFile string
	inPlace    bool

	rootCmd = &cobra.Command{
		Use:   "tf2bd_cleaner",
		Short: "Remove banned and deleted users from TF2 bot detector player lists",
		RunE: func(_ *cobra.Command, _ []string) error {
			if apiKey == "" {
				return errors.New("must set api key")
			}

			if err := steamweb.SetKey(apiKey); err != nil {
				return errors.Join(err, errors.New("could not set api key"))
			}

			return run()
		},
	}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

func run() error {
	var (
		ctx          = context.Background()
		reader       io.Reader
		inFileHandle *os.File
	)

	if inPlace && inputFile == "" {
		return errors.New("cannot use overwrite option without a input file")
	}

	if inputFile == "" {
		reader = bufio.NewReader(os.Stdin)
	} else {
		inFile, err := os.Open(inputFile)
		if err != nil {
			return errors.Join(err, errors.New("could not open input file"))
		}
		reader = inFile
	}

	var list TF2BDSchema
	if err := json.NewDecoder(reader).Decode(&list); err != nil {
		return errors.Join(err, errors.New("failed to decode list"))
	}

	if inFileHandle != nil {
		if err := inFileHandle.Close(); err != nil {
			return errors.Join(err, errors.New("failed to close input file"))
		}
	}

	var knownIDs []steamid.SteamID //nolint:prealloc
	for _, player := range list.Players {
		sid := steamid.New(player.Steamid)
		if !sid.Valid() {
			slog.Warn("Got invalid steamid: %v", player.Steamid)

			continue
		}
		knownIDs = append(knownIDs, sid)
	}

	slog.Info("Running profile checks...")
	banned, deleted, errBanned := findBanned(ctx, knownIDs)
	if errBanned != nil {
		return errBanned
	}

	var toRemove steamid.Collection
	toRemove = append(toRemove, banned...)
	toRemove = append(toRemove, deleted...)

	var players []Players

	for _, knownPlayer := range list.Players {
		if !slices.Contains(toRemove, steamid.New(knownPlayer.Steamid)) {
			players = append(players, knownPlayer)
		}
	}

	list.Players = players

	if errWrite := writeList(list); errWrite != nil {
		return errors.Join(errWrite, errors.New("failed to write output"))
	}

	slog.Info("Stats",
		slog.Int("total", len(knownIDs)),
		slog.Int("banned", len(banned)),
		slog.Int("deleted", len(deleted)),
		slog.Int("kept", len(knownIDs)-(len(banned)+len(deleted))))

	return nil
}

func writeList(list TF2BDSchema) error {
	var writer io.Writer

	if inPlace || outputFile != "" {
		var outPath string
		if inPlace {
			outPath = inputFile
		} else {
			outPath = outputFile
		}

		outFile, err := os.OpenFile(outPath, os.O_RDWR|os.O_CREATE, 0o755)
		if err != nil {
			return errors.Join(err, errors.New("could not open input file"))
		}
		writer = outFile

		defer func() {
			if errClose := outFile.Close(); errClose != nil {
				slog.Error("failed to close output file", slog.String("error", errClose.Error()))
			}
		}()
	} else {
		writer = os.Stdout
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent(" ", "    ")
	if errWrite := encoder.Encode(list); errWrite != nil {
		return errors.Join(errWrite, errors.New("failed to encode new player list"))
	}

	return nil
}

// findDeleted looks for accounts that are deleted. These have to be checked separately because the vac ban api endpoint will
// still return info for deleted accounts.
func findDeleted(ctx context.Context, ids steamid.Collection) (steamid.Collection, error) {
	var deleted steamid.Collection
	fetched, err := steamweb.PlayerSummaries(ctx, ids)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to fetch summaries"))
	}

	for _, checked := range ids {
		del := true
		for _, player := range fetched {
			if checked == player.SteamID {
				del = false

				break
			}
		}
		if del {
			deleted = append(deleted, checked)
		}
	}

	return deleted, nil
}

func findBanned(ctx context.Context, knownIDs steamid.Collection) (steamid.Collection, steamid.Collection, error) {
	var (
		currentSet steamid.Collection
		banned     steamid.Collection
		deleted    steamid.Collection
	)

	for i := range len(knownIDs) {
		currentSet = append(currentSet, knownIDs[i])
		if len(currentSet) == 100 || i == len(knownIDs) {
			// Deleted accounts
			deletedIDs, errDeleted := findDeleted(ctx, currentSet)
			if errDeleted != nil {
				return nil, nil, errDeleted
			}

			deleted = append(deleted, deletedIDs...)

			// Accounts with game/vacs
			bans, errBans := steamweb.GetPlayerBans(ctx, currentSet)
			if errBans != nil {
				return nil, nil, errors.Join(errBans, errors.New("failed to query steam api"))
			}

			for _, checked := range currentSet {
				wasDeleted := true
				for _, ban := range bans {
					if ban.SteamID == checked {
						wasDeleted = false

						break
					}
				}
				if wasDeleted {
					deleted = append(deleted, checked)
				}
			}

			for _, ban := range bans {
				if ban.VACBanned || ban.NumberOfGameBans > 0 {
					banned = append(banned, ban.SteamID)
				}
			}

			currentSet = nil
		}
	}

	return banned, deleted, nil
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&apiKey, "apikey", "k", "", "Steam API Key")
	rootCmd.PersistentFlags().StringVarP(&inputFile, "input", "i", "", "Input player list path. If not defined, stdin will be used")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "Output player list path. If not defined, stdout will be used")
	rootCmd.PersistentFlags().BoolVarP(&inPlace, "overwrite", "r", false, "Overwrite the input file.")
}
