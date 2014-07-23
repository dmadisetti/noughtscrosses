package noughtscrosses
import(
    "noughtscrosses/game"
    "net/http"
    "log"
    "fmt"
    "time"
    "encoding/json"
    "html/template"
    "strconv"
    "appengine"
    "appengine/datastore"
)

var t *template.Template

// Json hack from some blog (will update if I find again)
type Response map[string]interface{}
func (r Response) String() (s string) {
        b, err := json.Marshal(r)
        if err != nil {
                s = ""
                return
        }
        s = string(b)
        log.Println(s)
        return
}
// Struct to match move requests for post
type Move struct {
    Move byte
    Token string
}

// Start er up!
func init(){
    http.HandleFunc("/", mainHandle)
    http.HandleFunc("/post", postHandle)
    http.HandleFunc("/clean", cleanHandle)
}

// Handles
func mainHandle(w http.ResponseWriter, r *http.Request){
    c := appengine.NewContext(r)
    t, e := template.ParseGlob("templates/the.html")
    if e != nil {
        fmt.Fprint(w, e)        
        return
    }
    game := game.Create(c)
    err := t.Execute(w, game)
    if err !=nil{
        panic(err)
    }
}
func postHandle(w http.ResponseWriter, r *http.Request){
    // Check Post. Nothing we can do otherwise.
    if r.Method != "POST" {
        http.NotFound(w, r)
        return
    }

    // New context!
    c := appengine.NewContext(r)

    // Decode request
    decoder := json.NewDecoder(r.Body)
    var m Move
    err := decoder.Decode(&m)
    if err != nil {
        panic(err)
    }
    // Convert token
    token,err := strconv.ParseInt(m.Token,10,64)
    if err != nil {
        panic(err)
    }

    // Process and respond
    chosen, cat, lose, win := game.Process(c,token,m.Move)
    w.Header().Set("Content-Type", "application/json")
    switch true{
        case cat: 
            fmt.Fprint(w, Response{"message":"We drew"})
        case lose: 
            fmt.Fprint(w, Response{"message":"I lost"})
        case win: 
            fmt.Fprint(w, Response{"message":"You lost","index":chosen})
        // reads as normal
        default: fmt.Fprint(w, Response{"message":"","index":chosen})
    }
}
func cleanHandle(w http.ResponseWriter, r *http.Request){
    c := appengine.NewContext(r)
    message := "Thanks! Dones!"
    // Day old games
    q := datastore.NewQuery("Game").Filter("Date <=", time.Now().Add(-(time.Hour * 24)))
    keys, err := q.KeysOnly().GetAll(c,nil)
    if err == nil {
        err := datastore.DeleteMulti(c, keys)
        if err != nil{
            message = "Woops. My bad"
        }
    }else {
        message = "Broke on get."
    }
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprint(w,  Response{"message":message})

}
