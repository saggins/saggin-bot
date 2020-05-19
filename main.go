package main

import (
	"strings"
	"time"
	"fmt"
	"os"
	"log"
	"os/signal"
	"syscall"
	"encoding/json"
	"strconv"
	"io/ioutil"



	"github.com/bwmarrin/discordgo"
	"gopkg.in/yaml.v2"
	
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	"github.com/google/uuid"
)

type T struct {
	Ip 	string `yaml:"ip"`
	Token string `yaml:"token"`
	Channelid string `yaml:"channelid"`
	Port string `yaml:"port"`
}

var token string 
var ip string
var channelid string
var port string


func init(){

	token = os.Getenv("token")
	ip =os.Getenv("ip")
	channelid = os.Getenv("channelid")
	port = os.Getenv("port")

}

type status struct {
	Description chat.Message
	Players     struct {
		Max    int
		Online int
		Sample []struct {
			ID   uuid.UUID
			Name string
		}
	}
	Version struct {
		Name     string
		Protocol int
	}
	//favicon ignored
}

func main() {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}
	dg.AddHandler(ready)
	dg.AddHandler(guildCreate)
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}
	//the channel

	// nukes the channel
	
	//creates inital message
	msg, _, _:=  statusmsg() 
	chanMessages, _ :=  dg.ChannelMessages(channelid, int(30) ,"", "","" )
	for _, message  := range chanMessages {
		dg.ChannelMessageDelete(channelid, message.ID)
	}
	m, err := dg.ChannelMessageSend( channelid, msg)
	if (err!=nil ){
		fmt.Printf("channel msg send failure: %v", err)	
	}

	// Ticker to edit the message periodically
	go backgroundStatus(dg ,m.ID, m.ChannelID )

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Saggbot is ready .  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func backgroundStatus(s *discordgo.Session, messageid string, channelid string ) {

	ticker := time.NewTicker(1 * time.Second)
	for _ = range ticker.C {
		msg, up, player :=  statusmsg() 
		if (!player) {
			//pause := time.NewTicker(60 * time.Second)
		}
		if ( !up ) {
			//pause := time.NewTicker(60 * time.Second)

		}
		_, err:= s.ChannelMessageEdit( channelid  ,messageid, msg)
		if (err!= nil){
			fmt.Printf("Message edit error: %v", err)
		}
	}
}


func statusmsg() (string, bool, bool) {
	ports, _ := strconv.Atoi(port)
	resp, _, err := bot.PingAndList(ip, ports)
	if err != nil {
		msg  := "@Saggins#0250 :x: Sagg.in is Currently Down :(" 
		fmt.Printf("ping and list server fail: %v", err)
		return msg, false, false
	}

	var s status
	err = json.Unmarshal(resp, &s)

	msg := make([]string, len(s.Players.Sample))
	msg = append(msg, ":white_check_mark: : Sagg.in is up")
	for _, player  := range s.Players.Sample {
		playermsg := "-> " + player.Name
		msg = append(msg, playermsg)
	}
	leftover  := len(s.Players.Sample) - s.Players.Online
	msg =append(msg, "and "+ strconv.Itoa(leftover) +" more people")

	newmsg := strings.Join(msg, "\n")
	if (len(s.Players.Sample) == 0 ){
		return newmsg, true, false
	}
	return newmsg, true, true

}

func ready(s *discordgo.Session, event *discordgo.Ready){
	s.UpdateStatus(0, ip)
}
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {

	if event.Guild.Unavailable {
		return
	}

	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			_, _ = s.ChannelMessageSend(channel.ID, "Ey, bot is ready, just look at #minecraft-status")
			return
		}
	}
}