package main

import (
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/pat"
	"github.com/nats-io/nats.go"
	pb "github.com/victorolegovich/news-storage-service/proto"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func main() {
	router := pat.New()
	router.Get("/news/{id:[0-9A-z]+}", getNewsItemH)
	http.Handle("/", router)

	http.ListenAndServe("localhost:8080", nil)
}

func getNewsItemH(w http.ResponseWriter, r *http.Request) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		println("failed to create a logger zap")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("There were some internal server errors"))

		return
	}

	broker, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		logger.Error("Nats server connection error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("There were some internal server errors"))

		return
	}

	id := r.URL.Query().Get(":id")

	msg, err := broker.Request("storage", []byte(id), time.Second)
	if err != nil {
		logger.Error("message sending error news storage service", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("There were some internal server errors"))

		return
	}

	logger.Info("reply to the request was received", zap.String("subject", msg.Subject))

	newsItem := &pb.NewsItem{}
	if err = proto.Unmarshal(msg.Data, newsItem); err != nil {
		logger.Info("The news with this ID could not be found in the repository.", zap.String("id", id))
		w.WriteHeader(http.StatusNotFound)
		w.Write(msg.Data)
		return
	}

	jsontext := `{
	"id":"` + newsItem.ID + `",
	"header":"` + newsItem.Header + `",
}`

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte(jsontext))

	return
}
