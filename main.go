package noughtscrosses
import(
	"net/http"
	"log"
	"fmt"
    "bytes"
    "time"
	"encoding/json"
	"html/template"
    "strconv"
    "strings"
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

func init(){
    http.HandleFunc("/", mainHandle)
    http.HandleFunc("/clean", cleanHandle)
    http.HandleFunc("/post", postHandle)
}


func mainHandle(w http.ResponseWriter, r *http.Request){
    c := appengine.NewContext(r)
    t, e := template.ParseGlob("templates/index.html")
    if e != nil {
        fmt.Fprint(w, e)        
        return
    }
    game := newGame(c)
    err := t.Execute(w, game)
    if err !=nil{
        panic(err)
    }
}

func cleanHandle(w http.ResponseWriter, r *http.Request){
    c := appengine.NewContext(r)
    message := "Thanks! Dones!"
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

// Break out into proccessing library
// Not sure if this is a testimony to functional programming 
// ... or a bastardization of it
type directional struct{
    x, y iterator
    index func([]byte,int,int) byte
}
type iterator interface{
    iterate()
    getValue() int
}
type increment struct{value int}
func (inc *increment)getValue()int{return inc.value}
func (inc *increment)iterate(){
    if inc.value == getMax(){
        inc.value = getMin()
    }else{
        inc.value += 1
    }
}
type decrement struct{value int}
func (dec *decrement)getValue()int{return dec.value}
func (dec *decrement)iterate(){
    if dec.value == getMin(){
        dec.value = getMax()
    }else{
        dec.value -= 1
    }
}
func checkSymmetry(game []byte, direction *directional) (next []byte){
    next = make([]byte,9,9)
    for i := 0; i < 3; i++ {
        for j := 0; j < 3; j++ {
            next[i*3 + j] = direction.index(game,direction.x.getValue(),direction.y.getValue())
            direction.x.iterate()
        }
        direction.y.iterate()
    }
    return
}
func maximizeSwitch(game []byte, i int) []byte{
    switch i{
        case 1:return checkSymmetry(game, &directional{y:newIncrement(),x:newIncrement(),index:vertical}) // 0,3,6,1,4,7,2,5,8
        case 2:return checkSymmetry(game, &directional{y:newIncrement(),x:newDecrement(),index:vertical}) // 6,3,0,7,4,1,8,5,2
        case 3:return checkSymmetry(game, &directional{y:newDecrement(),x:newIncrement(),index:vertical}) // 2,5,8,1,4,7,0,3,6
        case 4:return checkSymmetry(game, &directional{y:newDecrement(),x:newDecrement(),index:vertical}) // 8,5,2,7,4,1,6,3,0
        case 5:return checkSymmetry(game, &directional{y:newIncrement(),x:newDecrement(),index:horizontal}) // 2,1,0,5,4,3,8,7,6
        case 6:return checkSymmetry(game, &directional{y:newDecrement(),x:newIncrement(),index:horizontal}) // 6,7,8,3,4,5,0,1,2
        case 7:return checkSymmetry(game, &directional{y:newDecrement(),x:newDecrement(),index:horizontal}) // 8,7,6,5,4,3,2,1,0
        // reads as normal
        default:return checkSymmetry(game, &directional{y:newIncrement(),x:newIncrement(),index:horizontal})
    }
}

func maximizeGame(game []byte) (max []byte){
    max = make([]byte,9)
    copy(max,game)
    for i := 1; i < 8; i++ {
        if next := maximizeSwitch(game,i); bytes.Compare(max,next) < 0{
            // next was larger
            copy(max,next)
            // use to be curious as to which transformation
            //which = i 
        }
    }
    return
}

// Helpers
func vertical(game []byte,x int,y int) byte{
    return game[x*3 + y]
}
func horizontal(game []byte,x int,y int) byte{
    return game[y*3 + x]
}
func newIncrement() *increment{
    return &increment{value:getMin()}
}
func newDecrement() *decrement{
    return &decrement{value:getMax()}
}
func getMax() int{
    return 2
}
func getMin() int{
    return 0
}
//////////////////////////////////////////////////////////////////////////////
type Record struct{
    FatalCount,Wins,Loses,Cats int
    Fatal bool
}

type Game struct{
    Plays []string
    Data []byte
    Id int
    Date time.Time
}

func newGame(c appengine.Context) Game{
    var game Game
    game.Data = make([]byte,9,9)
    game.Date = time.Now()

    key, err := datastore.Put(c,datastore.NewKey(c,"Game","",0, nil),&game)
    if err !=nil{
        panic(err)
    }
    game.Id = int(key.IntID())
    return game
}

func getGame(c appengine.Context, id int64) ([]string,[]byte){
    var game Game
    err := datastore.Get(c,datastore.NewKey(c,"Game","",id, nil),&game)
    if err !=nil{
        panic(err)
    }
    return game.Plays, game.Data
}

func setGame(c appengine.Context, plays []string, data []byte, id int64){
    var game Game
    key := datastore.NewKey(c,"Game","",id, nil)
    err := datastore.Get(c,key,&game)
    if err !=nil{
        panic(err)
    }
    game.Plays = plays
    game.Data = data
    datastore.Put(c,key,&game)
}

func (record *Record)getUtility(data []byte) int{
    if won(data,0x3){
        return 1024
    }else if record.Fatal{
        return -1024
    }else{
        // Arbitrary. Just keeps it from getting in a rut.
        return record.Wins * 2 + record.Cats - record.Loses * 2
    }
}

func lookUp(c appengine.Context, data []byte) (int, string){
    //database magic in here
    var record Record
    hash := string(data)
    err := datastore.Get(c,datastore.NewKey(c,"Record",hash,0, nil),&record)
    if err != nil{
        // creating a new enitity and relying on it's auto increment kills race condition errors 
        datastore.Put(c,datastore.NewKey(c,"Record",hash,0, nil),&record)
    }
    return record.getUtility(data),hash
}

var wins = [8][3]int{{0,1,2},{0,4,8},{0,3,6},{1,4,7},{2,4,6},{2,5,8},{3,4,5},{6,7,8}}

func won(game []byte, player byte) bool{
    tshift := player << player << player
    for i := 0; i < 8; i++ {
        if shift(game,wins[i]) == tshift{
            return true
        }
    }
    return false
}

func shift(game []byte,a [3]int) byte{
    return game[a[0]] << game[a[1]] << game[a[2]]
}

func recurse(c appengine.Context, state [3]int, id int64){
    var game Game
    err := datastore.Get(c,datastore.NewKey(c,"Game","",id, nil),&game)
    if err !=nil{
        log.Println(err)
        panic(err)
    }
    // Marked as lose
    var fatal = (state[1] == 1)
    // Check each record
    var record Record
    length := len(game.Plays) - 1
    for i := 0; i <= length; i++ {
        key := datastore.NewKey(c,"Record",game.Plays[length-i],0, nil)
        if err := datastore.Get(c,key,&record);err == nil{
            if fatal || record.FatalCount >= strings.Count(game.Plays[i],"\x00"){
                record.FatalCount += 1
                record.Fatal = true
                fatal = false
            }

            record.Wins  += state[2]
            record.Loses += state[1]
            record.Cats  += state[0]
            datastore.Put(c,key,&record)
        }else{
            log.Println(err)
            panic(err)
        }
    }
}

type Move struct {
    Move byte
    Token string
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
    token,err := strconv.ParseInt(m.Token,10,64)
    if err != nil {
        panic(err)
    }

    // Get game and set most recent move.
    game,data := getGame(c,token)
    data[m.Move] = 0x2

    // Out of scope
    var which string
    chosen  := 0
    lose := won(data,0x2)
    cat  := false
    win  := false
    if !lose{
        temp  := make([]byte,9)
        key   := make([]byte,9)
        maxUtility := -1025
        cat  = true
        for i := 0; i < 9; i++ {
            copy(temp,data);
            if temp[i] == 0{
                cat = false
                temp[i] = 0x3
                key = maximizeGame(temp)
                if utility,id:=lookUp(c, key);utility >= maxUtility{
                    chosen = i
                    which = id
                    maxUtility = utility
                }
            }
        }
        data[chosen] = 0x3
        game = append(game,which)
        win = won(data,0x3)
    }
    log.Println(data)
    log.Println(game)

    if !cat{
        setGame(c, game,data,token)
    }

    w.Header().Set("Content-Type", "application/json")
    switch true{
        case cat: 
            recurse(c, [3]int{1,0,0}, token)
            fmt.Fprint(w, Response{"message":"We drew"})
        case lose: 
            recurse(c, [3]int{0,1,0}, token)
            fmt.Fprint(w, Response{"message":"I lost"})
        case win: 
            recurse(c, [3]int{0,0,1}, token)
            fmt.Fprint(w, Response{"message":"You lost","index":chosen})
        // reads as normal
        default: fmt.Fprint(w, Response{"message":"","index":chosen})
    }
}