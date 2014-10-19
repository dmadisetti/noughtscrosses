package noughtscrosses

import(
    "time"
    "log"
    "strings"
    "appengine"
    "appengine/datastore"
)

// Record
type Record struct{
    FatalCount,Wins,Loses,Cats int
    Fatal bool
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
// Game
type Game struct{
    Plays []string
    Data []byte
    Id int
    Date time.Time
}

// Combinations of wins. Arrays are immutable though
var wins = [8][3]int{{0,1,2},{0,4,8},{0,3,6},{1,4,7},{2,4,6},{2,5,8},{3,4,5},{6,7,8}}

// Public
func Create(c appengine.Context) Game{
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
func Process(c appengine.Context, token int64, move byte)(chosen int, cat bool, lose bool, win bool){
    // Get game and set most recent move.
    game,data := getGame(c,token)
    if data[move] != 0x0 {
        // Die. Forged request.
        panic("Cheater.")
    }
    data[move] = 0x2

    // Out of scope
    var which string
    chosen  = 0
    lose = won(data,0x2)
    cat  = false
    win  = false
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
                key = Maximize(temp)
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
    switch true{
        case cat: 
            recurse(c, [3]int{1,0,0}, token)
        case lose: 
            recurse(c, [3]int{0,1,0}, token)
        case win: 
            recurse(c, [3]int{0,0,1}, token)
    }
    return chosen,cat,lose,win
}

// Private
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
// Grab game if exists
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
// Check if won by btye shift matching
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
// Game done, set records to reflect
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
