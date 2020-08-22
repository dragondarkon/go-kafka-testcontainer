package main

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

const (
	kafkaConn = "localhost:9092,broker:29092"
	topic     = "senz"
)

type Req struct {
	TopicName   string `json:"topic_name"`
	StartOffset int    `json:"start_offset"`
	EndOffset   int    `json:"end_offset"`
}

type Msg struct {
	Partition int32       `json:"partition"`
	Offset    int64       `json:"offset"`
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
}

type Res struct {
	TopicName  string `json:"topic_name"`
	TopicCount int    `json:"topic_count"`
	Msgs       []Msg  `json:"msg"`
}

func main() {

	// Echo instance
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// Routes
	e.GET("/", func(c echo.Context) error {
		msgs := TestConsume(topic, -1)
		res := Res{
			TopicName:  topic,
			TopicCount: len(msgs),
			Msgs:       msgs,
		}
		return c.JSON(http.StatusOK, res)
	})
	e.POST("/consume", func(c echo.Context) error {
		var req Req
		if err := c.Bind(&req); err != nil {
			return err
		}
		topicName := req.TopicName
		endOffset := req.EndOffset + 1
		msgs := TestConsume(topicName, endOffset)
		if msgs == nil {

			return c.JSON(http.StatusBadRequest, "consume over offset")
		}
		if len(msgs) < req.EndOffset-req.StartOffset {
			return c.JSON(http.StatusFound, "data comsume to less")
		}
		msgs = msgs[req.StartOffset:endOffset]
		res := Res{
			TopicName:  req.TopicName,
			TopicCount: req.EndOffset - req.StartOffset + 1,
			Msgs:       msgs,
		}
		return c.JSON(http.StatusOK, res)
	})
	// Start server
	e.Logger.Fatal(e.Start(":1323"))

}
