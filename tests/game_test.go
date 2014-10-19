package tests

import(
    "../noughtscrosses"
    "testing"
    "appengine/aetest"
    "appengine/datastore"
)

var c aetest.Context
var game noughtscrosses.Game
var test *testing.T
var id int64

func TestGame(t *testing.T){
    test = t
    context, err := aetest.NewContext(nil)
    if err != nil {
            test.Fatal(err)
    }
    c = context
    defer c.Close()
    
    // Create a game
    game = noughtscrosses.Create(c)
    test.Log("Game Created")

    // Convert id for lookup
    id = int64(game.Id)

    // Key for working with datastore
    key := datastore.NewKey(c,"Game","",id, nil)

    // Just move in the upper corner
    chosen, cat, lose, win := process(0, false)
    test.Logf("Init moved. Response %d", chosen)
    checkValid(chosen)
    checkNothing(cat, lose, win)

    // Try to move in the taken spot
    chosen, cat, lose, win = process(chosen, true)
    test.Log("Cannot move in same spot. Woot.")

    // Try to move in the original spot
    chosen, cat, lose, win = process(0, true)
    test.Log("Cannot move in same spot. Woot.")

    // Set winning game
    game.Data = []byte{0x3,0x3,0x3,
                  0x0,0x0,0x0,
                  0x0,0x0,0x0}
    datastore.Put(c,key,&game)
    chosen, cat, lose, win = process(4, false)
    checkWin(cat, lose, win)

    // Make sure I can lose
    game.Data = []byte{0x2,0x2,0x2,
                  0x0,0x0,0x0,
                  0x0,0x0,0x0}
    datastore.Put(c,key,&game)
    chosen, cat, lose, win = process(4, false)
    checkLose(cat, lose, win)

    // Make sure we can cat
    game.Data = []byte{0x3,0x2,0x3,
                  0x3,0x0,0x2,
                  0x2,0x3,0x2}
    datastore.Put(c,key,&game)
    chosen, cat, lose, win = process(4, false)
    checkCat(cat, lose, win)
}

func checkValid(chosen int){
    if chosen < 0 || chosen > 8 {
        test.Fatal("Returned invalid position")
    }
}

func checkNothing(cat bool, lose bool, win bool){
    if cat || lose || win{
        test.Fatal("Weird should have done nothing")
    }
    test.Log("Awesome. Nothing")
}

func checkWin(cat bool, lose bool, win bool){
    if cat || lose || !win{
        test.Fatal("Weird should have won")
    }
    test.Log("Awesome. Win")
}

func checkLose(cat bool, lose bool, win bool){
    if cat || !lose || win{
        test.Fatal("Weird should have lost")
    }
    test.Log("Awesome. Lose")
}

func checkCat(cat bool, lose bool, win bool){
    if !cat || lose || win{
        test.Fatal("Weird should have tied")
    }
    test.Log("Awesome. Cat")
}

func process(move int, invalid bool)(chosen int, cat bool, lose bool, win bool){
    defer func() {
        if r := recover(); r != nil {
            if(!invalid){
                test.Log(r)
                test.Fatal("Broke in process")
            }
        }else{
            if(invalid){
                test.Fatal("Should have broken in process")
            }
        }
    }()
    chosen, cat, lose, win = noughtscrosses.Process(c, id, byte(move))
    return
}