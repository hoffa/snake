package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/nsf/termbox-go"
)

const (
	Interval   = int64(50 * time.Millisecond)
	GrowAmount = 10
	FoodCount  = 5
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
	termbox.SetCell(c.x, c.y, ' ', termbox.ColorDefault, termbox.AttrReverse)
}

func NewSnake() *Snake {
	return &Snake{
		direction: Right,
		body:      make([]*Coord, 0),
		coords:    make(map[Coord]bool),
		grow:      GrowAmount,
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
	s.Push(&c)
	if s.grow > 0 {
		s.grow--
	} else {
		s.Pop()
	}
}

func (ctx *Context) Update() {
	ctx.Move(ctx.snake)
	ctx.AddFoods()
	for food := range ctx.foods {
		if ctx.snake.Occupies(&food) {
			ctx.snake.grow += GrowAmount
			delete(ctx.foods, food)
		}
	}
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for i, c := range strconv.Itoa(ctx.Score()) {
		termbox.SetCell(i, 0, c, termbox.ColorDefault, termbox.ColorDefault)
	}
	for food := range ctx.foods {
		food.Draw()
	}
	for _, c := range ctx.snake.body {
		c.Draw()
	}
	termbox.Flush()
}

func (ctx *Context) AddFoods() {
	w, h := termbox.Size()
	for len(ctx.foods) < FoodCount {
		food := Coord{rand.Intn(w - 1), rand.Intn(h - 1)}
		ctx.foods[food] = true
	}
}

func (ctx *Context) Score() int {
	return len(ctx.snake.body) - 1
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
	if err := termbox.Init(); err != nil {
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
	for !ctx.quit {
		select {
		case e := <-events:
			ctx.HandleKey(e.Key)
		default:
			now := time.Now().UnixNano()
			elapsed := now - ctx.updated
			if elapsed >= Interval {
				ctx.Update()
				ctx.updated = now
			}
			time.Sleep(time.Duration(Interval - elapsed))
		}
	}

	fmt.Println("Game over! Your score is", ctx.Score())
}
