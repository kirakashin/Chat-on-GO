package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type message struct {
	Time string
	Name string
	Body string
}

//all messages
var DATA []message

//all users
var USER []string

//login check

func login(r *http.Request) bool {
	for _, c := range r.Cookies() {
		if c.Name == "login" {
			return true
		}
	}
	return false
}

//getting value from cookie

func cookieValue(r *http.Request) string {
	for _, c := range r.Cookies() {
		if c.Name == "login" {
			return c.Value
		}
	}
	return ""
}

//getting index from user slice with a name as same as in login cookie

func nameIndex(r *http.Request) int {
	for _, c := range r.Cookies() {
		if c.Name == "login" {
			for i, v := range USER {
				if c.Value == v {
					return i
				}
			}
		}
	}
	return -1
}

//deleting name in USER

func nameDelete(r *http.Request) {
	var t []string
	USER = append(append(t, USER[:nameIndex(r)]...), USER[nameIndex(r)+1:]...)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("login.gohtml"))
	if r.Method != http.MethodPost {
		tmpl.Execute(w, nil)
		return
	}
	var name string = r.FormValue("Name")
	//check on already existing cookie
	if login(r) {
		if nameIndex(r) == -1 {
			http.SetCookie(w, &http.Cookie{
				Name:   "login",
				MaxAge: -1,
			})
		} else {
			nameDelete(r)
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:  "login",
		Value: name,
	})
	USER = append(USER, name)

	tmpl.Execute(w, struct{ Success bool }{true})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	//check on already existing cokies which are not in USER

	tmpl := template.Must(template.ParseFiles("logout.gohtml"))
	if login(r) {
		if nameIndex(r) != -1 {
			nameDelete(r)
		}
		http.SetCookie(w, &http.Cookie{
			Name:   "login",
			MaxAge: -1,
		})
		tmpl.Execute(w, Flag{true, true})
	} else {
		tmpl.Execute(w, nil)
	}

	if login(r) {
		if nameIndex(r) != -1 {
			nameDelete(r)
		}
		http.SetCookie(w, &http.Cookie{
			Name:   "login",
			MaxAge: -1,
		})
		fmt.Fprint(w, "Logout complete!")
	} else {
		fmt.Fprint(w, "Login first!")
	}
}

//for html
type Flag struct {
	Login bool
	Send  bool
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<a href=\"/\" style=\"margin-right: 20px;\">CHAT!</a>")
	if login(r) {
		if nameIndex(r) == -1 {
			fmt.Fprintf(w, "<a href=\"/login\" style=\"margin-right: 20px;\">Relogin!</a>")
		} else {
			fmt.Fprintf(w, "<a href=\"/send\" style=\"margin-right: 20px;\">Send a message!</a>")
			fmt.Fprintf(w, "<a href=\"/login\" style=\"margin-right: 20px;\">Rename!</a>")
		}
		fmt.Fprintf(w, "<a href=\"/logout\" style=\"margin-right: 20px;\">Logout!</a>")
	} else {
		fmt.Fprintf(w, "<a href=\"/login\" style=\"margin-right: 20px;\">Login!</a>")
	}
	fmt.Fprintf(w, "<a href=\"/count\" style=\"margin-right: 20px;\">Count!</a>")
	fmt.Fprintf(w, "<br>")
	for _, v := range DATA {
		fmt.Fprintf(w, "%[1]v\t[%[2]v]\t%[3]v\n", v.Time, v.Name, v.Body)
		fmt.Fprintf(w, "<br>")
	}
}
func sendHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("send.gohtml"))
	if r.Method != http.MethodPost {
		tmpl.Execute(w, nil)
		return
	}
	if login(r) {
		t := message{time.Now().Format("02-01 15:04:05"), cookieValue(r), r.FormValue("message")}
		DATA = append(DATA, t)
		tmpl.Execute(w, Flag{true, true})
		return
	}
	tmpl.Execute(w, Flag{false, true})
}

func countHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<a href=\"/\" style=\"margin-right: 20px;\">CHAT!</a>")
	if login(r) {
		fmt.Fprintf(w, "<a href=\"/send\" style=\"margin-right: 20px;\">Send a message!</a>")
		if nameIndex(r) == -1 {
			fmt.Fprintf(w, "<a href=\"/login\" style=\"margin-right: 20px;\">Relogin!</a>")
		} else {
			fmt.Fprintf(w, "<a href=\"/login\" style=\"margin-right: 20px;\">Rename!</a>")
		}
		fmt.Fprintf(w, "<a href=\"/logout\" style=\"margin-right: 20px;\">Logout!</a>")
	} else {
		fmt.Fprintf(w, "<a href=\"/login\" style=\"margin-right: 20px;\">Login!</a>")
	}
	fmt.Fprintf(w, "<a href=\"/count\" style=\"margin-right: 20px;\">Count!</a>")
	fmt.Fprintf(w, "<br>")
	fmt.Fprintf(w, "Number of logins: %v\n", len(USER))
	fmt.Fprintf(w, "<br>")
	fmt.Fprintf(w, "Online: %v\n", USER)

}

type Port struct {
	Port string
}

func main() {
	fmt.Println("Server up!")
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/count", countHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/", chatHandler)
	http.HandleFunc("/send", sendHandler)
	jsonFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	var port Port
	byte, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byte, &port)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Running on" + port.Port)
	err = http.ListenAndServe(port.Port, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}
