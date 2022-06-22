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
	x     float32
	y     float32
	rad   float32
	name  string
	color Color
}

type Food struct {
	x     uint16
	y     uint16
	rad   float32
	color Color
}

const (
	width  int = 1440
	height int = 900
	hz         = 60.0
)

const (
	playerRad float32 = 25.0
	foodRad   float32 = 7.0
)

var players []Player
var numPlayers byte = 0
var status byte = 0 // TODO: make const enum

var food []Food

const numFood byte = 50

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

	initMap()

	// Listen...
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		go handleClient(c)
	}
}

// initFood creates the initial state of the game.
func initMap() {
	food = make([]Food, numFood, numFood)
	for i := range food {
		food[i] = newFood()
	}
	go handleFood()
}

// newFood creates a new food item with a random position on the screen.
func newFood() Food {
	x, y := randCirclePos(foodRad)
	color := randColor()
	return Food{uint16(x), uint16(y), foodRad, color}
}

// handleFood checks whether players have collided with food and spawns new food.
func handleFood() { // TODO: make concurrent with splitting up the food list into 5 or similar.
	for {
		for i, f := range food {
			for j, p := range players {
				distX := p.x - float32(f.x)
				distY := p.y - float32(f.y)
				distance := math.Sqrt(float64(distX*distX) + float64(distY*distY))

				if distance+float64(f.rad) <= float64(p.rad) {
					food[i] = newFood()
					players[j].rad = combinedRad(float64(p.rad), float64(f.rad)) // TODO: see if data race is an issue.
				}
			}
		}
	}
}

// calcNewRad calculates the radius of a circle that has the same area as two other circles combined.
func combinedRad(rad1, rad2 float64) float32 {
	pi := math.Pi
	area := (pi * rad1 * rad1) + (pi * rad2 * rad2)
	return float32(math.Sqrt(area / pi))
}

// handleClient serves a single client connection.
func handleClient(c net.Conn) {
	defer c.Close()
	log.Printf("Serving %s\n", c.RemoteAddr().String())

	_, idx, err := initPlayer(bufio.NewReader(c))
	if err != nil {
		fmt.Printf("Read error: client (%s) disconnected, %s", c.RemoteAddr().String(), err)
		return
	}

	updateTime := time.Second / hz
	for {
		start := time.Now()

		write(c)
		msg, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Printf("Read error: client (%s) disconnected, %s", c.RemoteAddr().String(), err)
			removePlayer(idx)
			break
		}
		go handleMessage(msg, idx)

		elapsed := time.Since(start)
		fmt.Println("sleep: ", updateTime-elapsed)
		time.Sleep(updateTime - elapsed)
	}
}

// initPlayer reads in player information from client and creates a new player.
func initPlayer(rd *bufio.Reader) (Player, int, error) {
	color, err := readColor(rd)
	if err != nil {
		log.Println("Read error: color")
		return Player{}, 0, err
	}
	name, err := rd.ReadString('\n')
	if err != nil {
		log.Println("Read error: name")
		return Player{}, 0, err
	}

	x, y := randCirclePos(playerRad)
	player := Player{float32(x), float32(y), playerRad, name, color}

	if players == nil {
		players = make([]Player, 4, 10)
	}
	posFound := false
	idx := 0
	for i, p := range players {
		if p == (Player{}) {
			players[i] = player
			posFound = true
			idx = i
			break
		}
	}
	if !posFound {
		players = append(players, player)
		idx = len(players) - 1
	}

	numPlayers++
	return player, idx, nil
}

// write writes game data to the client.
func write(c net.Conn) {
	binary.Write(c, binary.LittleEndian, status)
	binary.Write(c, binary.LittleEndian, numFood)
	for _, f := range food {
		writePos(c, uint16(f.x), uint16(f.y))
		writeColor(c, f.color)
		writeRad(c, f.rad)
	}
	binary.Write(c, binary.LittleEndian, numPlayers)
	for _, p := range players { // TODO: what happens if a player is removed during the writing fase? i.e player count is sent, then a player is removed.
		if p != (Player{}) {
			writePos(c, uint16(p.x), uint16(p.y))
			writeColor(c, p.color)
			writeRad(c, p.rad)
			writeString(c, p.name)
		}
	}
}

// writePos writes an (x, y) position to the client.
func writePos(c net.Conn, x, y uint16) {
	binary.Write(c, binary.LittleEndian, x)
	binary.Write(c, binary.LittleEndian, y)
}

// writeColor writes an RGB color to the client.
func writeColor(c net.Conn, color Color) {
	binary.Write(c, binary.LittleEndian, color.r)
	binary.Write(c, binary.LittleEndian, color.g)
	binary.Write(c, binary.LittleEndian, color.b)
}

// writeRadius writes a circle radius to the client.
func writeRad(c net.Conn, r float32) {
	binary.Write(c, binary.LittleEndian, r)
}

// writeString writes a newline-ended string to the client.
func writeString(c net.Conn, s string) {
	c.Write([]byte(s))
}

// readColor reads in three bytes (red, green, blue) from client and
// returns them in that order. Error is returned if read is unsuccessful.
func readColor(rd *bufio.Reader) (Color, error) {
	r, err := rd.ReadByte()
	if err != nil {
		log.Println("Error: red")
		return Color{}, err
	}
	g, err := rd.ReadByte()
	if err != nil {
		log.Println("Error: green")
		return Color{}, err
	}
	b, err := rd.ReadByte()
	if err != nil {
		log.Println("Error: blue")
		return Color{}, err
	}
	return Color{r, g, b}, nil
}

// handleMessage processes the message from the client and moves player according to inputs.
func handleMessage(msg string, idx int) {
	msg = strings.TrimSpace(string(msg))
	vel := velFromRad(players[idx].rad)
	var dx float32 = 0
	var dy float32 = 0
	for _, v := range msg {
		switch string(v) {
		case "l":
			dx -= vel
		case "r":
			dx += vel
		case "u":
			dy -= vel
		case "d":
			dy += vel
		default:
			log.Println("Client sent invalid character!")
		}
	}
	movePlayer(idx, dx, dy)
}

// movePlayer moves a player and checks collisions with walls.
func movePlayer(idx int, dx, dy float32) {
	newX := players[idx].x + dx // TODO: solve potential data race
	newY := players[idx].y + dy // TODO: solve potential data race
	rad := players[idx].rad     // TODO: solve potential data race
	if newX-rad < 0 {
		newX = rad
	} else if newX+rad > float32(width) {
		newX = float32(width) - rad
	}
	if newY-rad < 0 {
		newY = rad
	} else if newY+rad > float32(height) {
		newY = float32(height) - rad
	}
	players[idx].x = newX // TODO: solve potential data race
	players[idx].y = newY // TODO: solve potential data race
}

// removePlayer removes a player from the game.
func removePlayer(idx int) {
	players[idx] = Player{}
	numPlayers--
}

// randColor returns a random RGB color.
func randColor() Color {
	return Color{byte(rand.Intn(255)), byte(rand.Intn(255)), byte(rand.Intn(255))}
}

// velFromRad calculates a players velocity based on it's radius.
func velFromRad(rad float32) float32 {
	return (0.0000385987 * rad * rad) - (0.0291013 * rad) + 5.73834
}

// randCirclePos finds a random spawning position for a player where no other player currently is.
func randCirclePos(rad float32) (x, y int) {
	for {
		x = rand.Intn(width)
		y = rand.Intn(height)

		if circleWallCollision(x, y, rad) {
			continue
		}

		found := false
		for _, p := range players {
			if p != (Player{}) {
				distX := x - int(p.x)
				distY := y - int(p.y)
				distance := math.Sqrt(float64(distX*distX) + float64(distY*distY))

				if float32(distance) <= (rad + p.rad) { // circles collide
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

// circleWallCollision checks if x, y with the default radius is within the screens boundaries.
func circleWallCollision(x, y int, rad float32) bool {
	if x-int(rad) < 0 {
		return true
	} else if x+int(rad) > width {
		return true
	} else if y-int(rad) < 0 {
		return true
	} else if y+int(rad) > height {
		return true
	}
	return false
}
