package subscribe

import (
	// "log"
	"fmt"
	"github.com/streadway/amqp"
	"sync"
	"errors"
)

type RabbitMQProducer struct {
	  Url string
		QueueName string
		Connections  []*amqp.Connection
		Channels []*amqp.Channel
		Current int
		mutex sync.Mutex
}

func NewRabbitMQProducer(url string, queue string, max int) *RabbitMQProducer {
	i := 0 ;
	conns := []*amqp.Connection{}
	for i < max {
			conn, err := amqp.Dial(url)
			if err != nil {
				fmt.Println(i, "Cannot connect to rabbit mq server: ", err)
				i = i + 1
				continue
			}
			conns = append(conns, conn)
			i = i + 1
	}
	chs := []*amqp.Channel{}
 	for _, conn := range conns {
		ch, err := conn.Channel()
		if err != nil {
			fmt.Println("Cannot connect create channel: ", err)
			continue
		}
		chs = append(chs,ch)
	}
	usable_chs := []*amqp.Channel{}
  for _, channel := range chs {
			if _, err := channel.QueueDeclare(
				queue, // name
				false,   // durable
				false,   // delete when unused
				false,   // exclusive
				false,   // no-wait
				nil,     // arguments
			); err != nil {
				fmt.Println("Cannot create queue: ",queue,", Error: ", err)
				continue
			}
			//Add channel to usable
			usable_chs = append(usable_chs, channel)
	}
	 rbmq_queue := &RabbitMQProducer{
		 Url: url,
		 QueueName: queue,
		 Connections: conns,
		 Channels: usable_chs,
		 Current: 0,
	 }
	 return rbmq_queue
}

func (q *RabbitMQProducer) Shutdown() bool {
		if q.Channels != nil {
			for _,channel := range q.Channels {
				 channel.Close()
			}
		}
		if q.Connections != nil {
			for _,conn := range q.Connections {
				 conn.Close()
			}
		}
		return true
}

func (q *RabbitMQProducer) GetChannel()  *amqp.Channel {
		q.mutex.Lock()
		defer q.mutex.Unlock()
		lch := len(q.Channels)
		if lch  == 0 {
			return nil
		}
		chanel := q.Channels[q.Current]
		q.Current = (q.Current + 1) % lch
		return chanel
}
func (q *RabbitMQProducer) Publish(msg string) error {
	channel := q.GetChannel()
	if channel == nil {
		return errors.New("Cannot get channel")
	}
	return channel.Publish(
		"",     // exchange
		q.QueueName, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		})
}
