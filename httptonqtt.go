package main

import (
    "fmt"
    "os"
    "log"
    "net/http"
    "github.com/gorilla/schema"
    "strconv"
    "encoding/json"
    "flag"

    MQTT "github.com/eclipse/paho.mqtt.golang"
)

// Http post structure
type Payload struct{
    Slot_temp string
    Data      string
    Time      string
    Device    string
    Signal    string
}

// Broker configuration
type MqttConfig struct {
    Host        string
    Port        int
    ClientID    string
    TopicPrefix string
    TopicSuffix string
    Qos         int
    Cleansess   bool
}

// Http server configuration
type HttpConfig struct {
    Host        string
    Port        int
}

type Configuration struct {
    Broker MqttConfig
    Http   HttpConfig
}

var (
    configFilename string
    config         Configuration
)

func postHandler(w http.ResponseWriter, r *http.Request) {

    err := r.ParseForm()

    if err != nil {
        log.Println("error:", err)
    }

    decoder := schema.NewDecoder()

    p := new(Payload)
    err = decoder.Decode(p, r.Form)

    if err != nil {
        log.Println("error:", err)
    }

    fmt.Println("\nReceived POST: ")
    fmt.Println(p)

    // Building MQTT topic and payling from received HTTP POST
    topic := p.Device
    payload := p.Data

    broker := "tcp://" + config.Broker.Host + ":" + strconv.Itoa(config.Broker.Port)
    topic = config.Broker.TopicPrefix + "/" + topic + "/" + config.Broker.TopicSuffix

    fmt.Printf("\nSample Info:\n")
    fmt.Printf("\tbroker:    %s\n", broker)
    fmt.Printf("\tclientId:  %s\n", config.Broker.ClientID)
    fmt.Printf("\ttopic:     %s\n", topic)
    fmt.Printf("\tmessage:   %s\n", payload)
    fmt.Printf("\tqos:       %d\n", config.Broker.Qos)
    fmt.Printf("\tcleansess: %v\n", config.Broker.Cleansess)

     //Set client options
    opts := MQTT.NewClientOptions()
    opts.AddBroker(broker)
    opts.SetClientID(config.Broker.ClientID)
    opts.SetCleanSession(config.Broker.Cleansess)

    client := MQTT.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }

    payload = fmt.Sprintf("{\"temp\": %s}", payload)

    // Publishing
    token := client.Publish(topic, 0, false, payload)
    token.Wait()

    client.Disconnect(250)
    fmt.Println("\nPublished and disconnected")
    fmt.Println(payload)
}

func main() {

    // Processing optional config file
    flag.StringVar(&configFilename, "config", "./config.json", "default value ./config.json")
    flag.Parse()

    log.Print("Using Configuration filename : " + configFilename)

    // Parsing configuration
    file, err := os.Open(configFilename)
    if err != nil {
        log.Println("error:", err)
        os.Exit(1)
    }
    decoder := json.NewDecoder(file)

    err = decoder.Decode(&config)
    if err != nil {
        log.Println("error:", err)
    }

    // Launching http server
    mux := http.NewServeMux()
    mux.HandleFunc("/", postHandler)
    log.Fatal(http.ListenAndServe(":8080", mux))
}
