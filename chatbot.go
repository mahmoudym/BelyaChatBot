package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"
	"os"
	"net/url"
  "io/ioutil"

	cors "github.com/heppu/simple-cors"
)

import _ "github.com/joho/godotenv/autoload"

var (
	// WelcomeMessage A constant to hold the welcome message
	WelcomeMessage = "Hello. I am Belya The Mechanic aka Mohamed Henedy, What's wrong with your car?"

	// sessions = {
	//   "uuid1" = Session{...},
	//   ...
	// }
	sessions = map[string]Session{}

	processor = sampleProcessor
)

type (
	// Session Holds info about a session
	Session map[string]interface{}

	// JSON Holds a JSON object
	JSON map[string]interface{}

	// Processor Alias for Process func
	Processor func(session Session, message string) (string, error)
)

func sampleProcessor(session Session, message string) (string, error) {
	// Make sure a history key is defined in the session which points to a slice of strings
	

	// Fetch the history from session and cast it to an array of strings
	_, statusfound := session["status"]
	
	
	if !statusfound {
		session["status"] = 1
	}
	
	// Make sure the message is unique in history
	
	
	if session["status"]==1 {
		if strings.EqualFold(strings.ToLower(message), "no") {
		session["status"] = 2
		}else{
			if strings.EqualFold(strings.ToLower(message), "yes"){
				return fmt.Sprintf("What else is wrong with your car ?"), nil
			}else{
				var p string = ""
	var d string = ""
	
	if strings.Contains(strings.ToLower(message), "steering"){
		p= "Steering"
		if strings.Contains(strings.ToLower(message), "left"){
			d="Left"
		}else{
			if strings.Contains(strings.ToLower(message), "right"){
				d = "Right"
			}else{
				d="Heavy"
			}
		}
		
	}
	
	
  resp,_ := http.PostForm("https://protected-brushlands-50304.herokuapp.com/", url.Values{"pro":{p}, "description":{d}})
	
	
  res,_ := ioutil.ReadAll(resp.Body)
  var raw map[string] interface{}
  json.Unmarshal([]byte(res),&raw)
	// Add the message in the parsed body to the messages in the session

	// Form a sentence out of the history in the form Message 1, Message 2, and Message 3
		if raw["Diagnosis"] !=nil{
				return fmt.Sprintf("Diagnosis: %s, " +"Steps To Solve The Problem: %s, Do you have any other problem ?",raw["Diagnosis"].(string), raw["Steps"].(string)), nil

		}else{
				return "", fmt.Errorf("I don't know what you're talking about")
		}
			}
		}
	

	}
	
	if session["status"] ==2 {
		if strings.EqualFold(strings.ToLower(message), "no") && session["Address"]==1 {
		session["status"] = 4
		}else{
			if strings.EqualFold(strings.ToLower(message), "yes") && session["Address"]==1{
				session["status"] =1
			}else{
			session["status"] = 3
			return fmt.Sprintf("What is your car manufacturer?"), nil
			}
		}
	}
	if session["status"] ==3 {
		var m string = ""
		if strings.Contains(strings.ToLower(message),"chevrolet"){
			m = "Chevrolet"
		}
		 resp,_ := http.PostForm("https://protected-brushlands-50304.herokuapp.com/tawkeel", url.Values{"man":{m}})
	
	
  res,_ := ioutil.ReadAll(resp.Body)
  var raw map[string] interface{}
  json.Unmarshal([]byte(res),&raw)
		if raw["Address"] !=nil {
	// Add the message in the parsed body to the messages in the session

	// Form a sentence out of the history in the form Message 1, Message 2, and Message 3
		session["Address"] = 1
		session["status"] = 2
		return fmt.Sprintf("Your service center address is: %s. is there another problem with your car ?",raw["Address"].(string)), nil
		}else{
			return "", fmt.Errorf("I can't find your car manufacturer please try again")
		}

	}
	
	if(session["status"]==4){
				return fmt.Sprintf("Thanks for using Belya the Mechanic, bye"), nil

	}
return fmt.Sprintf("What else is wrong with your car ?"), nil
}
				


// withLog Wraps HandlerFuncs to log requests to Stdout
func withLog(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := httptest.NewRecorder()
		fn(c, r)
		log.Printf("[%d] %-4s %s\n", c.Code, r.Method, r.URL.Path)

		for k, v := range c.HeaderMap {
			w.Header()[k] = v
		}
		w.WriteHeader(c.Code)
		c.Body.WriteTo(w)
	}
}

// writeJSON Writes the JSON equivilant for data into ResponseWriter w
func writeJSON(w http.ResponseWriter, data JSON) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// ProcessFunc Sets the processor of the chatbot
func ProcessFunc(p Processor) {
	processor = p
}

// handleWelcome Handles /welcome and responds with a welcome message and a generated UUID
func handleWelcome(w http.ResponseWriter, r *http.Request) {
	// Generate a UUID.
	hasher := md5.New()
	hasher.Write([]byte(strconv.FormatInt(time.Now().Unix(), 10)))
	uuid := hex.EncodeToString(hasher.Sum(nil))

	// Create a session for this UUID
	sessions[uuid] = Session{}

	// Write a JSON containg the welcome message and the generated UUID
	writeJSON(w, JSON{
		"uuid":    uuid,
		"message": WelcomeMessage,
	})
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	// Make sure only POST requests are handled
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed.", http.StatusMethodNotAllowed)
		return
	}

	// Make sure a UUID exists in the Authorization header
	uuid := r.Header.Get("Authorization")
	if uuid == "" {
		http.Error(w, "Missing or empty Authorization header.", http.StatusUnauthorized)
		return
	}

	// Make sure a session exists for the extracted UUID
	session, sessionFound := sessions[uuid]
	if !sessionFound {
		http.Error(w, fmt.Sprintf("No session found for: %v.", uuid), http.StatusUnauthorized)
		return
	}

	// Parse the JSON string in the body of the request
	data := JSON{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, fmt.Sprintf("Couldn't decode JSON: %v.", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Make sure a message key is defined in the body of the request
	_, problemmessages := data["message"]
	if !problemmessages {
		http.Error(w, "Missing message key in body.", http.StatusBadRequest)
		return
	}
	
	 

	// Process the received message
	message, err := processor(session, data["message"].(string))
	if err != nil {
		http.Error(w, err.Error(), 422  /* http.StatusUnprocessableEntity */ )
		return
	}
	// Write a JSON containg the processed response
	writeJSON(w, JSON{
		"message": message,
	})
}

// handle Handles /
func handle(w http.ResponseWriter, r *http.Request) {
	body :=
		"<!DOCTYPE html><html><head><title>Chatbot</title></head><body><pre style=\"font-family: monospace;\">\n" +
			"Available Routes:\n\n" +
			"  GET  /welcome -> handleWelcome\n" +
			"  POST /chat    -> handleChat\n" +
			"  GET  /        -> handle        (current)\n" +
			"</pre></body></html>"
	w.Header().Add("Content-Type", "text/html")
	fmt.Fprintln(w, body)
}

// Engage Gives control to the chatbot
func Engage(addr string) error {
	// HandleFuncs
	mux := http.NewServeMux()
	mux.HandleFunc("/welcome", withLog(handleWelcome))
	mux.HandleFunc("/chat", withLog(handleChat))
	mux.HandleFunc("/", withLog(handle))

	// Start the server
	return http.ListenAndServe(addr, cors.CORS(mux))
}

func main() {
	// Uncomment the following lines to customize the chatbot
	// chatbot.WelcomeMessage = "What's your name?"
	// chatbot.ProcessFunc(chatbotProcess)

	// Use the PORT environment variable
	port := os.Getenv("PORT")
	// Default to 3000 if no PORT environment variable was defined
	if port == "" {
		port = "3000"
	}

	// Start the server
	fmt.Printf("Listening on port %s...\n", port)
	log.Fatalln(Engage(":" + port))
}

