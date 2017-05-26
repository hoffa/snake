package main

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/nsf/termbox-go"
)

const (
	Interval   = 50 * time.Millisecond
	SnakeColor = termbox.ColorGreen
	FoodColor  = termbox.ColorRed
	GrowAmount = 10
	ScoreStep  = 1000
)

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

func NewCoord(x, y int) *Coord {
	return &Coord{x: x, y: y}
}

func (c *Coord) Draw(color termbox.Attribute) {
	termbox.SetCell(c.x, c.y, ' ', termbox.ColorDefault, color)
}

func (s *Snake) Push(c *Coord) {
	s.body = append(s.body, c)
	s.coords[c] = true
}

func (s *Snake) Pop() {
	delete(s.coords, s.body[0])
	s.body = s.body[1:]
}

func (s *Snake) Head() *Coord {
	return s.body[len(s.body)-1]
}

func (s *Snake) Occupies(c *Coord) bool {
	return s.coords[c]
}

type Snake struct {
	direction Direction
	body      []*Coord
	coords    map[*Coord]bool
	grow      int // Amount left to grow
}

func NewSnake(x, y int) *Snake {
	var body []*Coord
	coords := make(map[*Coord]bool)
	head := NewCoord(x, y)
	body = append(body, head)
	coords[head] = true
	return &Snake{
		direction: Up,
		body:      body,
		coords:    coords,
		grow:      10,
	}
}

type Context struct {
	quit  bool
	score int
	snake *Snake
	food  *Coord
}

func (s *Snake) Grow() {
	head := s.Head()
	switch s.direction {
	case Up:
		s.Push(NewCoord(head.x, head.y-1))
	case Down:
		s.Push(NewCoord(head.x, head.y+1))
	case Left:
		s.Push(NewCoord(head.x-1, head.y))
	case Right:
		s.Push(NewCoord(head.x+1, head.y))
	}
}

func (s *Snake) Move() {
	s.Grow()
	if s.grow <= 0 {
		s.Pop()
	} else {
		s.grow--
	}
}

func NewContext() *Context {
	w, h := termbox.Size()
	return &Context{
		snake: NewSnake(w/2, h/2),
		food:  NewCoord(-1, -1),
	}
}

func (ctx *Context) Draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	PrintInt(0, 0, ctx.score, termbox.ColorWhite)
	ctx.food.Draw(FoodColor)
	ctx.snake.Draw()
	termbox.Flush()
}

func (ctx *Context) Update() {
	// also how to check if ate itself? can't check head
	// cant go against itself
	if ctx.snake.Occupies(ctx.food) {
		ctx.score += ScoreStep
		ctx.snake.grow += GrowAmount
		ctx.RespawnFood()
	}
}

func (ctx *Context) RespawnFood() {
	w, h := termbox.Size()
	ctx.food.x = Random(0, w-1)
	ctx.food.y = Random(0, h-1)
}

func (s *Snake) Draw() {
	for _, c := range s.body {
		c.Draw(SnakeColor)
	}
}

func PrintInt(x, y, val int, color termbox.Attribute) {
	for i, c := range strconv.Itoa(val) {
		termbox.SetCell(x+i, y, c, color, termbox.ColorDefault)
	}
}

func Random(min, max int) int {
	return rand.Intn(max-min) + min
}

func (ctx *Context) HandleKey(key termbox.Key) {
	switch key {
	case termbox.KeyArrowUp:
		ctx.snake.direction = Up
	case termbox.KeyArrowDown:
		ctx.snake.direction = Down
	case termbox.KeyArrowLeft:
		ctx.snake.direction = Left
	case termbox.KeyArrowRight:
		ctx.snake.direction = Right
	}
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	events := make(chan termbox.Event)
	go func() {
		for {
			events <- termbox.PollEvent()
		}
	}()

	ctx := NewContext()
	rand.Seed(time.Now().Unix())
	ctx.RespawnFood()

	for !ctx.quit {
		select {
		case e := <-events:
			switch e.Type {
			case termbox.EventKey:
				if e.Key == termbox.KeyEsc {
					ctx.quit = true
				} else {
					ctx.HandleKey(e.Key)
				}
			}
		default:
			ctx.snake.Move()
			ctx.Update()
			ctx.Draw()
			time.Sleep(Interval)
		}
	}
}
