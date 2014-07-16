import 'dart:html';
import 'dart:async';
import 'dart:collection';
import 'dart:convert';

abstract class Player {

    Player(this.style);
    String style;
    StreamController moveController = new StreamController.broadcast();
    Stream get moved => moveController.stream;
}

class Human extends Player {
    Human(String style) :super(style);

    move(int i){
        this.moveController.add(i);
    }

    allow(event){

    }

}

class Ai extends Player {

    HttpRequest ask;
    String token = querySelector('.XOs').dataset['game'];
    StreamController gameController = new StreamController.broadcast();
    Stream get gameover => gameController.stream;

    Ai(String style) :super(style){
        prepareQuestion();
    }

    prepareQuestion(){
        ask = new HttpRequest();
        ask.open("POST", "/post");
        ask.onLoadEnd.listen(success);        
    }

    move(i){
        ask.send(JSON.encode({"token":token,"move":i}));
    }

    success(data){
        var answer = JSON.decode(ask.responseText);
        if(answer["index"] != null){
            this.moveController.add(answer["index"]);
        }
        if(answer["message"] == ""){
            prepareQuestion();
        }else{
            this.gameController.add(answer);
        }
    }
}

class Square{

    Element el;
    StreamController clickController = new StreamController.broadcast();
    Stream get clicked => clickController.stream;
    int i;
    var sub;

    Square(el,i){
        this.sub = el.onClick.listen(click);
        this.el = el;
        this.i = i;
    }

    void mark(Player p) {
        // Kills clickable this way too
        this.el.className = "${p.style} square";
        die();
    }

    void click(e){
        this.clickController.add(this);
    }

    void die(){
        this.sub.cancel();
    }
}

class Board<Square> extends ListBase<Square> {

    // Overrides to cheat
    final List<Square> l = [];
    void set length(int newLength) { l.length = newLength; }
    int get length => l.length;
    Square operator [](int index) => l[index];
    void operator []=(int index, Square value) { l[index] = value; }

    Human human;
    Ai    ai;   

    Board(Human this.human,Ai this.ai) : super(){
        this.human.moved.listen(ai.move);
        this.ai.moved.listen(human.allow);
        this.human.moved.listen(markHuman);
        this.ai.moved.listen(markAi);
        this.ai.gameover.listen(markDone);
    }

    markHuman(i){
        l[i].mark(this.human);
    }

    markAi(i){
        l[i].mark(this.ai);
    }

    markDone(message){
        for(var done in l){
            l[i].die();
        }
        alert(message);
    }

    add(square){
        l.add(square);
        square.clicked.listen(clicked);
    }

    clicked(Square square){
        this.human.move(square.i);
    }
}

main() {
    var human  = new Human('x');
    var ai     = new Ai('o');
    var board  = new Board(human,ai);
    {
        var i =0;
        for(var xo in querySelectorAll('.square')){
            var square = new Square(xo,i);
            board.add(square);
            i+=1;
        }
    }
}
