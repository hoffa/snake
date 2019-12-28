package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/nsf/termbox-go"
)

var interval int64
var growAmount int
var foodCount int
var color int
var party bool

type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

type Coord struct {
	x, y int
}

type Snake struct {
	direction Direction
	body      []*Coord
	coords    map[Coord]bool
	grow      int // Amount left to grow
}

type Context struct {
	quit    bool
	snake   *Snake
	foods   map[Coord]bool
	updated int64
}

func (c *Coord) Draw() {
	if party {
		termbox.SetCell(c.x, c.y, ' ', termbox.Attribute((color%6)+2), termbox.AttrReverse)
		color++
	} else {
		termbox.SetCell(c.x, c.y, ' ', termbox.Attribute(color), termbox.AttrReverse)
	}
}

func NewSnake() *Snake {
	return &Snake{
		direction: Right,
		body:      make([]*Coord, 0),
		coords:    make(map[Coord]bool),
		grow:      growAmount,
	}
}

func (s *Snake) Push(c *Coord) {
	s.body = append(s.body, c)
	s.coords[*c] = true
}

func (s *Snake) Pop() {
	delete(s.coords, *s.body[0])
	s.body = s.body[1:]
}

func (s *Snake) Occupies(c *Coord) bool {
	return s.coords[*c]
}

func NewContext() *Context {
	snake := NewSnake()
	snake.Push(&Coord{})
	return &Context{
		snake: snake,
		foods: make(map[Coord]bool),
	}
}

func (ctx *Context) Move(s *Snake) {
	w, h := termbox.Size()
	c := *s.body[len(s.body)-1]
	switch s.direction {
	case Down:
		c.y = (c.y + 1) % h
	case Right:
		c.x = (c.x + 1) % w
	case Up:
		c.y--
		if c.y < 0 {
			c.y += h
		}
	case Left:
		c.x--
		if c.x < 0 {
			c.x += w
		}
	}
	if s.Occupies(&c) {
		ctx.quit = true
	}
	c.Draw()
	s.Push(&c)
	if s.grow > 0 {
		s.grow--
	} else {
		// Clear last cell
		tail := *s.body[0]
		termbox.SetCell(tail.x, tail.y, ' ', termbox.ColorDefault, termbox.ColorDefault)
		s.Pop()
	}
}

func (ctx *Context) Update() {
	ctx.Move(ctx.snake)
	for food := range ctx.foods {
		if ctx.snake.Occupies(&food) {
			ctx.snake.grow += growAmount
			delete(ctx.foods, food)
		}
	}
	w, h := termbox.Size()
	for len(ctx.foods) < foodCount {
		food := Coord{rand.Intn(w - 1), rand.Intn(h - 1)}
		food.Draw()
		ctx.foods[food] = true
	}
	for i, c := range strconv.Itoa(ctx.Score()) {
		termbox.SetCell(i, 0, c, termbox.ColorDefault, termbox.ColorDefault)
	}
	termbox.Flush()
}

func (ctx *Context) Score() int {
	return len(ctx.snake.body) - 1
}

func (ctx *Context) HandleKey(key termbox.Key) {
	switch key {
	case termbox.KeyEsc:
		ctx.quit = true
	case termbox.KeyArrowUp:
		if ctx.snake.direction != Down {
			ctx.snake.direction = Up
		}
	case termbox.KeyArrowDown:
		if ctx.snake.direction != Up {
			ctx.snake.direction = Down
		}
	case termbox.KeyArrowLeft:
		if ctx.snake.direction != Right {
			ctx.snake.direction = Left
		}
	case termbox.KeyArrowRight:
		if ctx.snake.direction != Left {
			ctx.snake.direction = Right
		}
	}
}

func main() {
	speed := flag.Int64("speed", 9, "speed [0, 10]")
	flag.IntVar(&growAmount, "grow", 10, "grow amount per food")
	flag.IntVar(&foodCount, "food", 5, "foods on screen")
	flag.IntVar(&color, "color", 0, "color [0-9]")
	flag.BoolVar(&party, "party", false, "enable party mode")
	flag.Parse()
	interval = int64(250-(250/10)*(*speed-1)) * int64(time.Millisecond)

	if err := termbox.Init(); err != nil {
		panic(err)
	}

	events := make(chan termbox.Event)
	go func() {
		for {
			events <- termbox.PollEvent()
		}
	}()

	ctx := NewContext()
	rand.Seed(time.Now().Unix())
	for !ctx.quit {
		select {
		case e := <-events:
			ctx.HandleKey(e.Key)
		default:
			now := time.Now().UnixNano()
			elapsed := now - ctx.updated
			if elapsed >= interval {
				ctx.Update()
				ctx.updated = now
			}
			time.Sleep(time.Duration(interval - elapsed))
		}
	}

	termbox.Close()
	fmt.Println("Game over! Your score is", ctx.Score())
}
