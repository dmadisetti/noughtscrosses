package noughtscrosses

import("bytes")

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
// incrementor
type increment struct{value int}
func (inc *increment)getValue()int{return inc.value}
func (inc *increment)iterate(){
    if inc.value == max{
        inc.value = min
    }else{
        inc.value += 1
    }
}
// decrementor
type decrement struct{value int}
func (dec *decrement)getValue()int{return dec.value}
func (dec *decrement)iterate(){
    if dec.value == min{
        dec.value = max
    }else{
        dec.value -= 1
    }
}

// Yon Constants
const max int = 2
const min int = 0

// Public
func Maximize(game []byte) (max []byte){
    max = make([]byte,9)
    copy(max,game)
    for i := 1; i < 8; i++ {
        if next := maximizeSwitch(game,i); bytes.Compare(max,next) < 0{
            // next was larger
            copy(max,next)
            // use to be curious as to which transformation
        }
    }
    return
}

//Private
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

// Helpers
func vertical(game []byte,x int,y int) byte{
    return game[x*3 + y]
}
func horizontal(game []byte,x int,y int) byte{
    return game[y*3 + x]
}
func newIncrement() *increment{
    return &increment{value:min}
}
func newDecrement() *decrement{
    return &decrement{value:max}
}