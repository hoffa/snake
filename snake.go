package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/nsf/termbox-go"
)

const (
	Interval     = 25 * time.Millisecond
	SnakeColor   = termbox.ColorGreen
	FoodColor    = termbox.ColorRed
	GrowAmount   = 10
	ScoreStep    = 10
	VerticalSkip = 1 // no need for more than 1?
	FoodCount    = 10
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

func (c *Coord) Draw(color termbox.Attribute) {
	termbox.SetCell(c.x, c.y, ' ', termbox.ColorDefault, color)
}

func (s *Snake) Push(c *Coord) {
	s.body = append(s.body, c)
	s.coords[*c] = true
}

func (s *Snake) Pop() {
	delete(s.coords, *s.body[0])
	s.body = s.body[1:]
}

func (s *Snake) Head() *Coord {
	return s.body[len(s.body)-1]
}

func (s *Snake) Occupies(c *Coord) bool {
	return s.coords[*c]
}

type Snake struct {
	direction Direction
	body      []*Coord
	coords    map[Coord]bool
	grow      int // Amount left to grow
}

func NewSnake(x, y int) *Snake {
	snake := &Snake{
		direction: Up,
		body:      make([]*Coord, 0),
		coords:    make(map[Coord]bool),
		grow:      GrowAmount,
	}
	snake.Push(&Coord{x, y})
	return snake
}

type Context struct {
	quit         bool
	score        int
	snake        *Snake
	foods        map[Coord]bool
	verticalStep int
}

func (s *Snake) Grow(ctx *Context) {
	w, h := termbox.Size()
	head := s.Head()
	c := &Coord{head.x, head.y}
	switch s.direction {
	case Up:
		c.y--
		if c.y < 0 {
			c.y = h - 1
		}
	case Down:
		c.y++
		if c.y >= h {
			c.y = 0
		}
	case Left:
		c.x--
		if c.x < 0 {
			c.x = w - 1
		}
	case Right:
		c.x++
		if c.x >= w {
			c.x = 0
		}
	}
	if s.Occupies(c) {
		ctx.quit = true
	} else {
		s.Push(c)
	}
}

func (s *Snake) Move(ctx *Context) {
	s.Grow(ctx)
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
		foods: make(map[Coord]bool, 0),
	}
}

func (ctx *Context) Draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	PrintInt(0, 0, ctx.score, termbox.ColorWhite)
	ctx.DrawFoods()
	ctx.snake.Draw()
	termbox.Flush()
}

func (ctx *Context) Update() {
	if ctx.snake.direction == Up || ctx.snake.direction == Down {
		ctx.verticalStep++
		if ctx.verticalStep <= VerticalSkip {
			return
		}
	}
	ctx.verticalStep = 0
	ctx.snake.Move(ctx)
	for food := range ctx.foods {
		if ctx.snake.Occupies(&food) {
			ctx.score += ScoreStep
			ctx.snake.grow += GrowAmount
			delete(ctx.foods, food)
			ctx.AddFood()
		}
	}
}

func (ctx *Context) AddFood() {
	w, h := termbox.Size()
	for len(ctx.foods) < FoodCount {
		food := Coord{Random(0, w-1), Random(0, h-1)}
		ctx.foods[food] = true
	}
}

func (ctx *Context) DrawFoods() {
	for food := range ctx.foods {
		food.Draw(FoodColor)
	}
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
	err := termbox.Init()
	if err != nil {
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
	ctx.AddFood()

	for !ctx.quit {
		ctx.Update()
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
			ctx.Draw()
			time.Sleep(Interval)
		}
	}

	termbox.Close()
	fmt.Println("Game over! Your score is", ctx.score)
}
