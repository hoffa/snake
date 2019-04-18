package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/nsf/termbox-go"
)

const (
	Speed      = 25 * time.Millisecond
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
	quit  bool
	snake *Snake
	foods map[Coord]bool
	skip  bool
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

func (ctx *Context) Grow(s *Snake) {
	w, h := termbox.Size()
	c := *s.body[len(s.body)-1]
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
	if s.Occupies(&c) {
		ctx.quit = true
	} else {
		s.Push(&c)
	}
}

func (ctx *Context) Move(s *Snake) {
	ctx.Grow(s)
	if s.grow <= 0 {
		s.Pop()
	} else {
		s.grow--
	}
}

func (ctx *Context) Draw() {
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

func (ctx *Context) Update() {
	if ctx.snake.direction == Up || ctx.snake.direction == Down {
		ctx.skip = !ctx.skip
		if ctx.skip {
			return
		}
	} else {
		ctx.skip = false
	}
	ctx.Move(ctx.snake)
	for food := range ctx.foods {
		if ctx.snake.Occupies(&food) {
			ctx.snake.grow += GrowAmount
			delete(ctx.foods, food)
		}
	}
	ctx.AddFoods()
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

func update(ctx *Context, events chan termbox.Event) {
	ctx.Update()
	select {
	case e := <-events:
		switch e.Type {
		case termbox.EventKey:
			ctx.HandleKey(e.Key)
		}
	default:
		ctx.Draw()
		time.Sleep(Speed)
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
		update(ctx, events)
	}

	fmt.Println("Game over! Your score is", ctx.Score())
}
