// dictionary - A maubot plugin to get dictionary word definitions.
// Copyright (C) 2018 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>. 

package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"maubot.xyz"
	"maubot.xyz/dictionary/source"
)

type DictionaryBot struct {
	client maubot.MatrixClient
	log    maubot.Logger

	dict source.Source
}

const CommandDictionary = "dictionary $word"

func (bot *DictionaryBot) Start() {
	//bot.dict = oxford.New(http.DefaultClient, "12a3b4c6", "1a2345bcd678h9e012fgh34567i89jkl")
	//bot.dict = webster.New(http.DefaultClient, "68453e05-312a-40ad-b809-a51a7498f76b")
	panic("Dictionary keys not set")
	bot.client.SetCommandSpec(&maubot.CommandSpec{
		Commands: []maubot.Command{{
			Syntax:      CommandDictionary,
			Description: "Get the dictionary definition(s) for a single word.",
			Arguments: maubot.ArgumentMap{
				"$word": {
					Matches:     "\\w+",
					Required:    true,
					Description: "The word to get the definition of.",
				},
			},
		}},
	})
	bot.client.AddCommandHandler(CommandDictionary, bot.CmdDictionaryHandler)
}

func (bot *DictionaryBot) Stop() {}

func (bot *DictionaryBot) FormatSenses(senses []source.Sense) string {
	if len(senses) == 0 {
		return ""
	}
	var str strings.Builder
	str.WriteString("<ol>")
	for _, sense := range senses {
		if len(sense.Definitions()) == 0 {
			d, _ := json.MarshalIndent(sense, "", "  ")
			bot.log.Debugln("Weird sense encountered:", string(d))
			continue
		}
		str.WriteString("<li>")
		str.WriteString(sense.Definitions()[0])
		if len(sense.Examples()) > 0 {
			str.WriteString("<br/>")
			for _, example := range sense.Examples() {
				fmt.Fprintf(&str, "<blockquote><em>%s</em></blockquote>", example)
			}
		}
		fmt.Fprint(&str, bot.FormatSenses(sense.Subsenses()))
		str.WriteString("</li>")
	}
	str.WriteString("</ol>")
	return str.String()
}

func (bot *DictionaryBot) FormatDefinition(result source.Result) string {
	var str strings.Builder
	fmt.Fprintf(&str, "### %s\n", result.Headword())
	for _, entry := range result.Entries() {
		fmt.Fprintln(&str, "***")
		fmt.Fprintf(&str, "**%s**\n", entry.Category())
		fmt.Fprintln(&str, bot.FormatSenses(entry.Senses()))
	}
	return str.String()
}

func (bot *DictionaryBot) CmdDictionaryHandler(evt *maubot.Event) maubot.CommandHandlerResult {
	word := evt.Content.Command.Arguments["$word"]
	if len(word) == 0 {
		return maubot.Continue
	}

	bot.log.Debugln("Fetching definition of", word, "from", bot.dict.Name(), "for", evt.Sender)
	def, err := bot.dict.Define(word)
	if res, notFound := err.(*source.EmptyResultError); notFound {
		evt.Reply("No definitions found for " + res.Word + " :(")
		return maubot.StopCommandPropagation
	} else if err != nil {
		evt.Reply("Failed to get definition: " + err.Error())
		return maubot.StopCommandPropagation
	}

	d, _ := json.MarshalIndent(def, "", "  ")
	bot.log.Debugln(string(d))
	_, err = evt.Reply(bot.FormatDefinition(def))
	if err != nil {
		bot.log.Debugln(err)
	}
	return maubot.StopCommandPropagation
}

var Plugin = maubot.PluginCreator{
	Create: func(client maubot.MatrixClient, logger maubot.Logger) maubot.Plugin {
		return &DictionaryBot{
			client: client,
			log:    logger,
		}
	},
	Name:    "maubot.xyz/dictionary",
	Version: "0.1.0",
}
