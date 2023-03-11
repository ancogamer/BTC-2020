package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/jeffotoni/gconcat"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unicode"
	"unicode/utf8"
)

//=======================
//Desafio da 7 da maratona behind the code 2020 :D
//https://github.com/maratonadev-br/desafio-7-2020
type Writer struct {
	Comma   rune // Field delimiter (set to ',' by NewWriter)
	UseCRLF bool // True to use \r\n as the line terminator
	w       *bufio.Writer
}
type dados struct {
	Row          int    `json:"row"`
	Tempo        string `json:"Tempo"`
	Estação      string `json: "Estação"`
	LAT          string `json: "LAT"`
	LONG         string `json:"LONG"`
	Movimentacao string `json: "Movimentação"`
	Original473  string `json:"Original_473"`
	Original269  string `json:"Original_269"`
	Zero         string `json:"Zero"`
	MacaVerde    string `json: "Maçã-Verde"`
	Tangerina    string `json:"Tangerina"`
	Citrus       string `json: "Citrus"`
	AcaiGuaraná  string `json: "Açaí-Guaraná"`
	Pessego      string `json:"Pêssego"`
	TARGET       string `json:"TARGET"`
}

// many thanks to :
// https://stackoverflow.com/questions/48872360/golang-mqtt-publish-and-subscribe
// https://golangcode.com/write-data-to-a-csv-file/
// https://golang.org/pkg/encoding/csv/
var knt int
var i int
var o int
var controle [17016]int
var t bool
var insert = [17016][14]string{{"Tempo", "Estação", "LAT", "LONG", "Movimentação", "Original_473", "Original_269", "Zero", "Maçã-Verde", "Tangerina", "Citrus", "Açaí-Guaraná", "Pêssego", "TARGET"}}
var f MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	j:=i
	//Tempo,Estação,LAT,LONG,Movimentação,Original_473,Original_269,Zero,Maçã-Verde,Tangerina,Citrus,Açaí-Guaraná,Pêssego,Status
	println(j)
	println(string(msg.Payload()))
	var dado dados
	err := json.Unmarshal(msg.Payload(), &dado)
	switch err {
	case nil:
	default:
		panic(gconcat.Build("erro:", err))
	}
	switch j {
	case 0:
		j++
		controle[0]=dado.Row
		insert[1][0] = dado.Tempo
		insert[1][1] = dado.Estação
		insert[1][2] = dado.LAT
		insert[1][3] = dado.LONG
		insert[1][4] = dado.Movimentacao
		insert[1][6] = dado.Original269
		insert[1][7] = dado.Zero
		insert[1][8] = dado.MacaVerde
		insert[1][9] = dado.Tangerina
		insert[1][10] = dado.Citrus
		insert[1][11] = dado.AcaiGuaraná
		insert[1][12] = dado.Pessego
		insert[1][13] = dado.TARGET
		break
	case 17016:
		writeCSV(insert)
	default:
		t = true
		for _,controle:= range controle{
		//for o=0 ;o<i;o++ {
			switch {
			case dado.Row == controle:
			/*
				println(i)
				println("/==============================")
				println(controle)
				println(dado.Row)
				println("/==============================")
			 */
				t=false
			}
		}
		switch  t {
		case true:
			controle[j]=dado.Row
			insert[j][0] = dado.Tempo
			insert[j][1] = dado.Estação
			insert[j][2] = dado.LAT
			insert[j][3] = dado.LONG
			insert[j][4] = dado.Movimentacao
			insert[j][5] = dado.Original473
			insert[j][6] = dado.Original269
			insert[j][7] = dado.Zero
			insert[j][8] = dado.MacaVerde
			insert[j][9] = dado.Tangerina
			insert[j][10] = dado.Citrus
			insert[j][11] = dado.AcaiGuaraná
			insert[j][12] = dado.Pessego
			insert[j][13] = dado.TARGET
			j++
		}
	}
	i=j

}



func writeCSV(s [17016][14]string) {
	println("\nCREATING FILE <3")
	file, err := os.Create("result.csv")
	checkError("Cannot create file", err)
	defer file.Close()

	writer := NewWriter(file)
	defer writer.Flush()
	for _, value := range s {
		err := writer.Write(value)
		checkError("Cannot write to file", err)
	}
}

//FROM THE CSV PACKAGE i just edit the Write func so i can pass the [14]string as value====================================================
func (w *Writer) Flush() {
	w.w.Flush()
}
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		Comma: ',',
		w:     bufio.NewWriter(w),
	}
}
func (w *Writer) Write(record [14]string) error {
	for n, field := range record {
		if n > 0 {
			if _, err := w.w.WriteRune(w.Comma); err != nil {
				return err
			}
		}

		// If we don't have to have a quoted field then just
		// write out the field and continue to the next field.
		if !w.fieldNeedsQuotes(field) {
			if _, err := w.w.WriteString(field); err != nil {
				return err
			}
			continue
		}

		if err := w.w.WriteByte('"'); err != nil {
			return err
		}
		for len(field) > 0 {
			// Search for special characters.
			i := strings.IndexAny(field, "\"\r\n")
			if i < 0 {
				i = len(field)
			}

			// Copy verbatim everything before the special character.
			if _, err := w.w.WriteString(field[:i]); err != nil {
				return err
			}
			field = field[i:]

			// Encode the special character.
			if len(field) > 0 {
				var err error
				switch field[0] {
				case '"':
					_, err = w.w.WriteString(`""`)
				case '\r':
					if !w.UseCRLF {
						err = w.w.WriteByte('\r')
					}
				case '\n':
					if w.UseCRLF {
						_, err = w.w.WriteString("\r\n")
					} else {
						err = w.w.WriteByte('\n')
					}
				}
				field = field[1:]
				if err != nil {
					return err
				}
			}
		}
		if err := w.w.WriteByte('"'); err != nil {
			return err
		}
	}
	var err error
	if w.UseCRLF {
		_, err = w.w.WriteString("\r\n")
	} else {
		err = w.w.WriteByte('\n')
	}
	return err
}
func (w *Writer) fieldNeedsQuotes(field string) bool {
	if field == "" {
		return false
	}

	if field == `\.` {
		return true
	}

	if w.Comma < utf8.RuneSelf {
		for i := 0; i < len(field); i++ {
			c := field[i]
			if c == '\n' || c == '\r' || c == '"' || c == byte(w.Comma) {
				return true
			}
		}
	} else {
		if strings.ContainsRune(field, w.Comma) || strings.ContainsAny(field, "\"\r\n") {
			return true
		}
	}

	r1, _ := utf8.DecodeRuneInString(field)
	return unicode.IsSpace(r1)
}

//=============================================================================================================================================================
func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
	knt++
}

func main() {
	knt = 0
	i = 0
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	opts := MQTT.NewClientOptions().AddBroker("mqtt://tnt-iot.maratona.dev:30573")
	opts.Password = "ndsjknvkdnvjsbvj"
	opts.Username = "maratoners"
	opts.SetClientID("macgfyh-go")
	opts.SetDefaultPublishHandler(f)


		opts.SetClientID(gconcat.Build("go é amor numero :"))
		opts.OnConnect = func(c MQTT.Client) {
			if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
				panic(token.Error())
			}
		}
		client := MQTT.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			panic(token.Error())
		} else {
			fmt.Printf("Connected to server\n")
		}
		opts.OnConnect = func(c MQTT.Client) {
			if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
				panic(token.Error())
			}
		}
	opts.SetClientID(gconcat.Build("go é amor numero :"))
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		fmt.Printf("Connected to server\n")
	}
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	opts.SetClientID(gconcat.Build("go é amor numero :"))
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		fmt.Printf("Connected to server\n")
	}
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	opts.SetClientID(gconcat.Build("go é amor numero :"))
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		fmt.Printf("Connected to server\n")
	}
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	opts.SetClientID(gconcat.Build("go é amor numero :"))
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		fmt.Printf("Connected to server\n")
	}
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	opts.SetClientID(gconcat.Build("go é amor numero :"))
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		fmt.Printf("Connected to server\n")
	}
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	opts.SetClientID(gconcat.Build("go é amor numero :"))
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		fmt.Printf("Connected to server\n")
	}
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	opts.SetClientID(gconcat.Build("go é amor numero :"))
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		fmt.Printf("Connected to server\n")
	}
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	opts.SetClientID(gconcat.Build("go é amor numero :"))
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		fmt.Printf("Connected to server\n")
	}
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	opts.SetClientID(gconcat.Build("go é amor numero :"))
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		fmt.Printf("Connected to server\n")
	}
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	opts.SetClientID(gconcat.Build("go é amor numero :"))
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		fmt.Printf("Connected to server\n")
	}
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	opts.SetClientID(gconcat.Build("go é amor numero :"))
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		fmt.Printf("Connected to server\n")
	}
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	opts.SetClientID(gconcat.Build("go é amor numero :"))
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		fmt.Printf("Connected to server\n")
	}
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	opts.SetClientID(gconcat.Build("go é amor numero :"))
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		fmt.Printf("Connected to server\n")
	}
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	opts.SetClientID(gconcat.Build("go é amor numero :"))
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		fmt.Printf("Connected to server\n")
	}
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("tnt", 0, f); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}






















	<-c

} //end
