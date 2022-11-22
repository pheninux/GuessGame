package main

import (
	"fmt"
	"github.com/matryer/runner"
	"math/rand"
	"strings"
	"time"
)

type player struct {
	id   int
	name string
	win  int
	turn bool
}

type task struct {
	t *runner.Task
}

var mots = []string{"PLOMBIER", "CAFE", "SUCRE", "BOITE"}
var endSubscribePlayer bool
var plz []player
var endGame bool = false
var rw string
var Player = 0
var c = make(chan runner.S)
var tasks []*runner.Task

func main() {

	mw := choosedWord(mots)

	i := 0
	pseudo := ""

	for !endGame {

		for !endSubscribePlayer {
			i++
			fmt.Printf("player n: %v , taper votre pseudo : \n", i)
			fmt.Scan(&pseudo)
			p := player{
				id:   i,
				name: pseudo,
				win:  0,
				turn: false,
			}
			plz = append(plz, p)

		mylabel:
			fmt.Println("Voulez vous ajouter un nouveau player ?")
			var sub string
			fmt.Scan(&sub)

			if sub != "no" && sub != "yes" {
				fmt.Println("reponse incorrect , taper yes ou no pour continuer")
				goto mylabel
			}

			switch sub {
			case "no":
				endSubscribePlayer = true
				fmt.Println("***************startCounDown GAME******************")
				fmt.Println("***************A VOS MARQUE****************")
				showWord(mw)
				//fmt.Printf("%s à avous :", plz[chooseTheStartPlayer(plz)].name)
			}
		}

		var res string

		for {

			showTurnPlayer()

			//we lunch go routine
			t := runner.Go(startCountDownTask)
			tasks = append(tasks, t)
			runner.Go(startCheckTask)
			// here we wait until player submit response
			fmt.Scan(&res)
			sp := Player // sotck submited player
			// here if player submit word verifaying if it mutch
			if strings.ToLower(res) == strings.ToLower(rw) {
				plz[sp].win++
				fmt.Println("bravooooooo :)")

				// here if we finich all word we should exit game
				if len(mots) == 0 {
					break
				}
				// show another world for players
				showWord(choosedWord(mots))

			}
			//here if we are at the end of player we initalize var player to 0
			if Player == (len(plz) - 1) {
				Player = 0
			} else {
				Player++
			}
			for _, t := range tasks {
				t.Stop()
			}
		}

		fmt.Println("########### Resulta final #####################")

		// finaly print the final score for all players
		for _, p := range plz {
			fmt.Printf("player :%s   score :%v\n", p.name, p.win)
		}

		endGame = true

	}
}

func showTurnPlayer() {
	fmt.Printf("%s à avous :\n", plz[Player].name)
	fmt.Println("compt a rebour activé 10 scd:")
}

func startCountDownTask(s runner.S) error {

	for i := 10; i > 0; i-- {
		time.Sleep(time.Second * 1)
	}
	c <- s
	return nil
}

func startCheckTask(s runner.S) error {

	for {
		select {
		case s := <-c:
			if !s() {
				Player++
				if len(plz) == Player {
					Player = 0
				}
				showTurnPlayer()
				t := runner.Go(startCountDownTask)
				tasks = append(tasks, t)
			}
		}
	}
	return nil
}

func choosedWord(m []string) []string {
	ran := rand.Intn(len(mots))
	rw = mots[ran]                             // rw => random world pickup
	mots = append(mots[:ran], mots[ran+1:]...) // delete the selected word
	sow := strings.Split(rw, "")               // sow => slice of caracters from word
	return MixWord2(sow)                       // mixed word
}

// reverse les valeur un slice
func Shuffle(words []string) []string {
	//r := rand.New(rand.NewSource(time.Now().Unix()))
	word := make([]string, len(words))
	n := len(words)
	for i := 0; i < n; i++ {
		randIndex := rand.Intn(len(words))
		word[i] = words[randIndex]
		words = append(words[:randIndex], words[randIndex+1:]...)
	}
	return word
}

func MixWord(words []string) []string {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	word := make([]string, len(words))
	n := len(words)
	for i := 0; i < n; i++ {
		randIndex := r.Intn(len(words))
		word[i] = words[randIndex]
		words = append(words[:randIndex], words[randIndex+1:]...)
	}
	return word
}

func MixWord2(word []string) []string {

	var res []string
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for _, i := range r.Perm(len(word)) {
		val := word[i]
		res = append(res, val)
	}
	return res
}

func chooseTheStartPlayer(p []player) int {

	r := rand.New(rand.NewSource(time.Now().Unix()))
	return r.Intn(len(p))
}

func chooseTheStartPlayer2(val int) int {

	r := rand.New(rand.NewSource(time.Now().Unix()))
	return r.Intn(val)
}

func showWord(wr []string) {
	fmt.Println("*******************************************")
	fmt.Println("Le mot est : ", wr)
	fmt.Println("*******************************************")
}
