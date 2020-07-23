package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"path/filepath"
)

type Challenge struct {
	ID          string   `json:"id"`               // A short memorable identifier for challenge, users will look challenges up with this
	Name        string   `json:"name"`             // Title should be a minimal hint towards the type of challenge this is
	Description string   `json:"description"`      // Description that contains the location of the challenge
	Hints       []string `json:"hints"`            // Hints in increasing order of helpfulness
	Flag        string   `json:"flag"`             // The final flag without any wrapper{text}
	Role        string   `json:"role"`             // Role should reflect a key to the BotConfig.Roles map
	FileName    string   `json:filename,omitempty` // Optional FileName to server with this challenge
	FileType    string   `json:filetype,omitempty` // If you specify a FileName a FileType (MIME type) must be provided
	Link        string   `json:link,omitempty`     // Option URL for the challenge message ot link to
	Color       string   `json:color,omitempty`    // Optional ability to specify the color of the embed, Black(0) is default
}
type BotConfig struct {
	DiscordToken   string            // Discord Bot token, must be specified in BOT_TOKEN environment var
	DiscordGuild   string            `json:"guild"`             // Guild ID for the bot to operate in
	DiscordChannel string            `json:"channel,omitempty"` // If you want the bot to only respond in one channel use this
	DefaultColor   string            `json:"default_color"`     // Default Color to use for the embed message if the challenge does not specify one
	CommandString  string            `json:"command_string"`    // CommandString represents what each command much start with
	Roles          map[string]string `json:"roles"`             // Map of readable role names used in the application to discord role ids
	Challenges     []Challenge       `json:"challenges"`        // Array of challenge objects
}

var config *BotConfig                        // Stores the global config for the Bot, should be accessed via Config()
var guildRoleNames = make(map[string]string) // A cache of the guild roles, this is only loaded on start up

func main() {
	cfg := Config()
	discord, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		panic(err)
	}
	if discord.Open() != nil {
		panic(err)
	}
	defer discord.Close()
	discord.AddHandler(messageCreate)

	// Initialize the Guild Role map so this doesn't need to be looked up every time there is a solve
	if roles, err := discord.GuildRoles(cfg.DiscordGuild); err == nil {
		for _, role := range roles {
			guildRoleNames[role.ID] = role.Name
		}
	} else {
		fmt.Printf("Unable to retrieve roles, are you sure the guild ID(%s) is correct?\n", cfg.DiscordGuild)
		return
	}

	fmt.Println("Running...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

// Wrapper for config access, loads from file on first access
func Config() BotConfig {
	if config == nil {
		c := BotConfig{}
		val := os.Getenv("BOT_TOKEN")
		if val != "" {
			c.DiscordToken = val
		} else {
			panic("No BOT_TOKEN provided")
		}

		val = os.Getenv("BOT_CONFIG_FILE")
		if val != "" {
			if file, err := ioutil.ReadFile(val); err != nil {
				panic(err)
			} else {
				if err = json.Unmarshal([]byte(file), &c); err != nil {
					panic(err)
				}
			}
		} else {
			panic("No BOT_CONFIG_FILE provided")
		}
		config = &c
	}
	return *config
}

// DiscordGo callback to handle new messages
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	ch, _ := s.Channel(m.ChannelID)
	if ch.Type == discordgo.ChannelTypeDM {
		start := strings.Index(m.Content, "{")
		end := strings.LastIndex(m.Content, "}")
		if start > -1 && end > start+2 {
			checkFlag(s, m, m.Content[start+1:end])
		}
	} else {
		channel := Config().DiscordChannel
		if channel != "" && channel != m.ChannelID {
			return
		}
	}

	if strings.HasPrefix(m.Content, Config().CommandString) {
		info := strings.SplitN(strings.Trim(m.Content[len(Config().CommandString):], " \t\n\r"), " ", 2)
		info[0] = strings.ToLower(info[0])

		switch info[0] {
		case "ping":
			s.ChannelMessageSend(m.ChannelID, "Pong!")
		case "challenges":
			commandListChallenges(s, m)
		case "challenge":
			commandDisplayChallenge(s, m, info)
		case "hint":
			commandHint(s, m, info)
		case "commands":
			commandHelp(s, m)
		}
	}
}

// Simple wrapper that create a new User Channel with a user and then sends a message to it
func sendDM(s *discordgo.Session, userId, message string) {
	if cid, err := s.UserChannelCreate(userId); err == nil {
		s.ChannelMessageSend(cid.ID, message)
	}
}

// Looks up a challenge by the id field
func getChallengeById(id string) (Challenge, error) {
	for _, v := range Config().Challenges {
		if v.ID == id {
			return v, nil
		}
	}
	return Challenge{}, errors.New(fmt.Sprintf("Could not find challenge with id '%s'\n", id))
}

// Command handler for listing all the available challenges
func commandListChallenges(s *discordgo.Session, m *discordgo.MessageCreate) {
	output := "```"
	for _, v := range Config().Challenges {
		output += fmt.Sprintf("%-18s %s\n", "["+v.ID+"]", v.Name)
	}
	output += "```"
	s.ChannelMessageSend(m.ChannelID, output)
}

// Command handler to display information about a specific challenge and serve and files necessary
func commandDisplayChallenge(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) != 2 {
		return
	}
	args[1] = strings.ToLower(args[1])

	challenge, err := getChallengeById(args[1])
	if err != nil {
		fmt.Print(err)
		return
	}

	message := discordgo.MessageSend{}
	message.Embed = &discordgo.MessageEmbed{}
	message.Embed.Title = challenge.Name
	message.Embed.Description = fmt.Sprintf("This Challenge unlocks the **%s** role.\n\n", guildRoleNames[Config().Roles[challenge.Role]])
	message.Embed.Description += challenge.Description

	if challenge.Color == "" {
		challenge.Color = Config().DefaultColor
	}
	if challenge.Color != "" && strings.ToLower(challenge.Color)[:2] == "0x" {
		if color, err := strconv.ParseUint(challenge.Color[2:], 16, 32); err == nil {
			message.Embed.Color = int(color)
		} else {
			fmt.Print(err)
		}
	}

	if challenge.Link != "" {
		message.Embed.URL = challenge.Link
	}

	if challenge.FileName != "" && challenge.FileType != "" {
		basePath, _ := filepath.Abs("./files")
		file := filepath.Join(basePath, challenge.FileName)

		if len(file) <= len(basePath) || file[:len(basePath)] != basePath {
			panic("Attempted to serve file from outside files directory.")
		}

		fd, err := os.Open(file)
		defer fd.Close()
		if err == nil {
			message.File = &discordgo.File{}
			message.File.Name = challenge.FileName
			message.File.ContentType = challenge.FileType
			message.File.Reader = fd
		} else {
			fmt.Printf("[!!] Could not open file(%s) - %s\n", challenge.FileName, err)
		}
	}
	s.ChannelMessageSendComplex(m.ChannelID, &message)
}

// If the bot receives a suspected flag the suspected flag is passed her for validation
func checkFlag(s *discordgo.Session, m *discordgo.MessageCreate, submittedFlag string) {
	for _, v := range Config().Challenges {
		if v.Flag == submittedFlag {
			var cfg = Config()
			var newRoleID = cfg.Roles[v.Role]
			var roleName = guildRoleNames[newRoleID]
			if roleName != "" {
				s.GuildMemberRoleAdd(cfg.DiscordGuild, m.Author.ID, newRoleID)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You've unlocked the %s role.", roleName))
			}
		}
	}
}

// Command handler to display hints (via DM) if a user requests a hint
func commandHint(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) != 2 {
		return
	}
	words := strings.Split(strings.ToLower(args[1]), " ")
	if challenge, err := getChallengeById(words[0]); err == nil {
		if len(challenge.Hints) == 0 {
			sendDM(s, m.Author.ID, fmt.Sprintf("[%s] Has no hints.", challenge.Name))
			return
		}

		var hID = 0
		if len(words) >= 2 {
			if hID, err = strconv.Atoi(words[1]); err == nil {
				hID -= 1
			} else {
				hID = 0
			}
		}

		if hID < 0 {
			hID = 0
		} else if hID >= len(challenge.Hints) {
			hID = len(challenge.Hints) - 1
		}

		hint := challenge.Hints[hID]
		sendDM(s, m.Author.ID, fmt.Sprintf("Hint %d of %d for %s\n||%s||", hID+1, len(challenge.Hints), challenge.Name, hint))
	}
}

// Command handler that displays the static help message
func commandHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	cstring := Config().CommandString
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```"+`The following commands are supported:
%schallenges - Will display a list of challenge IDs and their titles
%schallenge <id> - Will display all the information you need to know to get started on the challenge
%shint <id> [number] - If you do not provide a [number] the first hint for the challenge will be displayed
%scommands - This...

Roles can be obtained by submitting flags through a DM to this account.
To submit a flag send a DM containing only the flag in the form of FLAG{s0me_flag_h3re} to the bot

DO NOT for any reason attempt to submit a flag through any public chat. The flags are not to be shared.
`+"```", cstring, cstring, cstring, cstring))
}
