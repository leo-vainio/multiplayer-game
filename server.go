package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

type Color struct {
	r, g, b byte
}

type Player struct {
	x      uint16
	y      uint16
	radius float64
	name   string
	color  Color
}

const (
	width  int = 1440
	height int = 900
	hz         = 60.0
)

const (
	playerRad float64 = 10.0
	foodRad   float32 = 7.0
)

var players []Player
var numPlayers byte = 0
var status byte = 0

// main initializes server and starts listening for clients.
func main() {
	addr := flag.String("addr", "localhost:8080", "http service address")
	flag.Parse()

	log.SetFlags(0)
	log.Println("Starting server:", *addr)

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer ln.Close()

	for {
		c, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		go handleClient(c)
	}
}

// handleClient serves a single client connection.
func handleClient(c net.Conn) {
	defer c.Close()
	log.Printf("Serving %s\n", c.RemoteAddr().String())

	player, idx, err := initPlayer(bufio.NewReader(c))
	if err != nil {
		fmt.Printf("Read error: client (%s) disconnected, %s", c.RemoteAddr().String(), err)
		return
	}

	// TODO: remove
	fmt.Println(player, idx)

	update := time.Second / hz
	for {
		start := time.Now()

		// if player is nil do something: doesnt mean player has left necessarily.

		// ----- WRITE ----- //
		write(c)

		// ----- READ ----- //
		msg, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Printf("Read error: client (%s) disconnected, %s", c.RemoteAddr().String(), err)
			removePlayer(idx)
			break
		}

		// ----- HANDLE ----- //
		go func(msg string, idx int) { // TODO: move to separate function.
			if len(msg) > 6 {
				return
			}
			msg = strings.TrimSpace(string(msg))
			l, r, u, d, s := false, false, false, false, false
			for _, v := range msg {
				switch string(v) {
				case "l":
					l = true
				case "r":
					r = true
				case "u":
					u = true
				case "d":
					d = true
				case "s":
					s = true // TODO: implement
					fmt.Print(s)
				default:
					log.Println("Client sent invalid character!")
				}
			}
			vel := uint16(10)
			if l {
				go func() {
					players[idx].x -= vel
				}()
			}
			if r {
				go func() {
					players[idx].x += vel
				}()
			}
			if u {
				go func() {
					players[idx].y -= vel
				}()
			}
			if d {
				go func() {
					players[idx].y += vel
				}()
			}

		}(msg, idx)

		elapsed := time.Since(start)
		// fmt.Println("sleep: ", update-elapsed) // for testing efficiency TODO: remove
		time.Sleep(update - elapsed)
	}
}

// ballWallCollision checks if x, y with the default radius is within the screens boundaries.
func ballWallCollision(x, y int) bool {
	if x-int(playerRad) < 0 {
		return true
	} else if x+int(playerRad) > width {
		return true
	} else if y-int(playerRad) < 0 {
		return true
	} else if y+int(playerRad) > height {
		return true
	}
	return false
}

// randStartPos finds a random spawning position for a player where no other player currently is.
func randStartPos() (x, y int) {
	for {
		x = rand.Intn(width)
		y = rand.Intn(height)

		if ballWallCollision(x, y) {
			continue
		}

		found := false
		for _, p := range players {
			if p != (Player{}) {
				distX := x - int(p.x)
				distY := y - int(p.y)
				distance := math.Sqrt(float64(distX*distX) + float64(distY*distY))

				if distance <= (playerRad + p.radius) { // circles collide
					found = true
					break
				}
			}
		}
		if !found {
			break
		}
	}
	return
}

func initPlayer(rd *bufio.Reader) (Player, int, error) {
	r, g, b, err := readColor(rd)
	if err != nil {
		log.Println("Read error: color")
		return Player{}, 0, err
	}
	name, err := rd.ReadString('\n')
	if err != nil {
		log.Println("Read error: name")
		return Player{}, 0, err
	}

	x, y := randStartPos() // TODO: fix

	player := Player{x: uint16(x), y: uint16(y), radius: playerRad, name: name, color: Color{r, g, b}}

	if players == nil {
		players = make([]Player, 4, 10)
	}
	slotFound := false
	idx := 0
	for i, p := range players {
		if p == (Player{}) {
			players[i] = player
			slotFound = true
			idx = i
			break
		}
	}
	if !slotFound {
		players = append(players, player)
		idx = len(players) - 1
	}

	numPlayers++
	return player, idx, nil
}

// readColor reads in three bytes (red, green, blue) from client and
// returns them in that order. Error is returned if read is unsuccessful.
func readColor(rd *bufio.Reader) (byte, byte, byte, error) {
	r, err := rd.ReadByte()
	if err != nil {
		log.Println("Error: red")
		return 0, 0, 0, err
	}
	g, err := rd.ReadByte()
	if err != nil {
		log.Println("Error: green")
		return 0, 0, 0, err
	}
	b, err := rd.ReadByte()
	if err != nil {
		log.Println("Error: blue")
		return 0, 0, 0, err
	}
	return r, g, b, nil
}

// write writes game data to the client.
func write(c net.Conn) {
	binary.Write(c, binary.LittleEndian, status)     // status
	binary.Write(c, binary.LittleEndian, numPlayers) // player count

	for _, p := range players {
		if p != (Player{}) {
			binary.Write(c, binary.LittleEndian, p.x)               // x
			binary.Write(c, binary.LittleEndian, p.y)               // y
			binary.Write(c, binary.LittleEndian, p.color.r)         // red
			binary.Write(c, binary.LittleEndian, p.color.g)         // green
			binary.Write(c, binary.LittleEndian, p.color.b)         // blue
			binary.Write(c, binary.LittleEndian, float32(p.radius)) // radius
			c.Write([]byte(p.name))                                 // name
		}
	}
}

// removePlayer removes a player from the game.
func removePlayer(idx int) {
	players[idx] = Player{}
	numPlayers--
}
