package main

// ----------------------------------------------------------------------------
// IMPORTS
// ----------------------------------------------------------------------------
import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"
)

// ----------------------------------------------------------------------------
// TYPES
// ----------------------------------------------------------------------------
type Color int

type Tile struct {
	Value int
	Color Color
}

type Player struct {
	ID             int
	Hand           []Tile
	HasPlayedFirst bool
	Name           string
	IsAI           bool
}

type Combination []Tile

type GameState struct {
	Players           []Player
	Remaining         []Tile
	CurrentPlayerID   int
	Table             [][]Tile
	Hand              []Tile
	ConsecutivePasses int
}

// ----------------------------------------------------------------------------
// CONSTS
// ----------------------------------------------------------------------------
const (
	Red Color = iota
	Blue
	Green
	Orange
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorBlue   = "\033[34m"
	ColorOrange = "\033[33m"
)

// ----------------------------------------------------------------------------
// main()
// ----------------------------------------------------------------------------
func main() {
	args := os.Args
	if len(args) != 2 {
		fmt.Printf("Usage: %s <nombre_de_joueurs>\n", args[0])
		return
	}

	numPlayers := 0
	_, err := fmt.Sscanf(args[1], "%d", &numPlayers)
	if err != nil || numPlayers < 2 || numPlayers > 4 {
		fmt.Println("Le nombre de joueurs doit être entre 2 et 4.")
		return
	}
	game := InitializeGame(numPlayers)
	game.CurrentPlayerID = game.DetermineFirstPlayer()

	for {
		player := &game.Players[game.CurrentPlayerID]

		if player.IsAI {
			game.IATurn(player)
		} else {
			game.HumanTurn(player)
		}

		// On passe au joueur suivant
		game.CurrentPlayerID = (game.CurrentPlayerID + 1) % len(game.Players)
	}
}

// ----------------------------------------------------------------------------
// IsValidGroup()
// ----------------------------------------------------------------------------
func IsValidGroup(combo Combination) bool {
	if len(combo) < 3 || len(combo) > 4 {
		return false
	}

	firstValue := -1
	colorsSeen := make(map[Color]bool)

	for _, tile := range combo {
		// Gestion du Joker (valeur 0)
		if tile.Value == 0 {
			continue
		}

		// Vérification de la valeur unique (ex: tous des 11)
		if firstValue == -1 {
			firstValue = tile.Value
		} else if tile.Value != firstValue {
			return false // Valeurs différentes dans un groupe !
		}

		// --- LA CORRECTION EST ICI ---
		// Vérification des couleurs uniques
		if colorsSeen[tile.Color] {
			return false // Doublon de couleur détecté !
		}
		colorsSeen[tile.Color] = true
	}

	return true
}

// ----------------------------------------------------------------------------
// IsValidRun()
// ----------------------------------------------------------------------------
func IsValidRun(combo Combination) bool {
	if len(combo) < 3 {
		return false
	}

	// 1. Extraire les tuiles réelles et compter les jokers
	var realTiles []Tile
	jokerCount := 0
	var color Color
	colorSet := false

	for _, tile := range combo {
		if tile.Value == 0 {
			jokerCount++
		} else {
			if !colorSet {
				color = tile.Color
				colorSet = true
			} else if tile.Color != color {
				return false // Toutes les tuiles réelles doivent être de la même couleur
			}
			realTiles = append(realTiles, tile)
		}
	}

	// Si tout est Joker (cas rare mais possible), c'est valide
	if len(realTiles) == 0 {
		return true
	}

	// 2. Trier les tuiles réelles
	sort.Slice(realTiles, func(i, j int) bool {
		return realTiles[i].Value < realTiles[j].Value
	})

	// 3. Vérifier les doublons et les écarts
	for i := 0; i < len(realTiles)-1; i++ {
		diff := realTiles[i+1].Value - realTiles[i].Value

		if diff == 0 {
			return false // Deux tuiles identiques (ex: deux 7 rouges)
		}

		// Si l'écart est > 1, on consomme des jokers pour combler
		// Ex: entre 5 et 8, il faut 2 jokers (le 6 et le 7)
		neededJokers := diff - 1
		jokerCount -= neededJokers
	}

	// 4. Vérifier les limites du Rummikub (valeurs de 1 à 13)
	// On vérifie si les jokers restants peuvent être placés sans sortir des clous
	// (Cette partie est simplifiée ici, on vérifie surtout si jokerCount n'est pas négatif)
	if jokerCount < 0 {
		return false
	}

	// Optionnel : Vérifier que la suite totale (réelles + jokers utilisés)
	// ne dépasse pas 13 tuiles et reste entre 1 et 13.
	return len(combo) <= 13
}

// ----------------------------------------------------------------------------
// IsValidCombination()
// ----------------------------------------------------------------------------
func IsValidCombination(combo Combination) bool {
	return IsValidGroup(combo) || IsValidRun(combo)
}

// ----------------------------------------------------------------------------
// initializeAllTiles()
// ----------------------------------------------------------------------------
func initializeAllTiles() []Tile {
	var tiles []Tile
	// 2 sets of 52 tiles (4 colors x 13 values)
	for i := 0; i < 2; i++ {
		for value := 1; value <= 13; value++ {
			for color := Red; color <= Orange; color++ {
				tiles = append(tiles, Tile{Value: value, Color: color})
			}
		}
	}
	// Add 2 jokers
	tiles = append(tiles, Tile{Value: 0, Color: -1})
	tiles = append(tiles, Tile{Value: 0, Color: -1})
	return tiles
}

// ----------------------------------------------------------------------------
// shuffleTiles()
// ----------------------------------------------------------------------------
func shuffleTiles(tiles []Tile) []Tile {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(tiles), func(i, j int) {
		tiles[i], tiles[j] = tiles[j], tiles[i]
	})
	return tiles
}

// ----------------------------------------------------------------------------
// dealTiles()
// ----------------------------------------------------------------------------
func dealTiles(tiles []Tile, numPlayers int) ([]Player, []Tile) {
	players := make([]Player, numPlayers)
	for i := 0; i < numPlayers; i++ {
		players[i] = Player{ID: i + 1}
		var name string
		isAI := false

		if i == 0 {
			name = "Humain" // Le premier joueur, c'est toi
			isAI = false
		} else {
			// Les suivants sont des IA (AI#1, AI#2...)
			name = fmt.Sprintf("AI#%d", i)
			isAI = true
		}
		players[i].Name = name
		players[i].IsAI = isAI
		players[i].HasPlayedFirst = false
	}

	// Each player gets 14 tiles
	for i := 0; i < numPlayers*14; i++ {
		playerIndex := i / 14
		players[playerIndex].Hand = append(players[playerIndex].Hand, tiles[i])
	}

	// The rest forms the draw pile
	remaining := tiles[numPlayers*14:]
	return players, remaining
}

// ----------------------------------------------------------------------------
// InitializeGame()
// ----------------------------------------------------------------------------
func InitializeGame(numPlayers int) GameState {
	if numPlayers < 2 || numPlayers > 4 {
		panic("Le nombre de joueurs doit être entre 2 et 4.")
	}

	allTiles := initializeAllTiles()
	shuffledTiles := shuffleTiles(allTiles)
	players, remaining := dealTiles(shuffledTiles, numPlayers)

	return GameState{
		Players:         players,
		Remaining:       remaining,
		Table:           [][]Tile{},
		CurrentPlayerID: 0,
	}
}

// ----------------------------------------------------------------------------
// SortTiles()
// ----------------------------------------------------------------------------
func SortTiles(tiles []Tile) {
	sort.Slice(tiles, func(i, j int) bool {
		// 1. Comparer les couleurs
		if tiles[i].Color != tiles[j].Color {
			return tiles[i].Color < tiles[j].Color
		}
		// 2. Si les couleurs sont identiques, comparer les valeurs
		return tiles[i].Value < tiles[j].Value
	})
}

// ----------------------------------------------------------------------------
// removeTiles()
// ----------------------------------------------------------------------------
func removeTiles(hand []Tile, combo Combination) []Tile {
	newHand := make([]Tile, len(hand))
	copy(newHand, hand)

	for _, comboTile := range combo {
		for i, handTile := range newHand {
			// On vérifie la valeur et la couleur (le Joker a Value: 0)
			if handTile.Value == comboTile.Value && handTile.Color == comboTile.Color {
				newHand = append(newHand[:i], newHand[i+1:]...)
				break
			}
		}
	}
	return newHand
}

// ----------------------------------------------------------------------------
// FindBestHandLayout()
// ----------------------------------------------------------------------------
func FindBestHandLayout(hand []Tile) ([][]Tile, []Tile) {
	var bestTable [][]Tile
	minTilesLeft := len(hand)
	bestTable = [][]Tile{}
	remainingHand := hand

	// Fonction récursive interne
	var backtrack func(currentHand []Tile, currentTable [][]Tile)
	backtrack = func(currentHand []Tile, currentTable [][]Tile) {
		// On cherche toutes les combinaisons possibles dans la main actuelle
		// possibilities := append(FindAllRuns(currentHand), FindAllGroups(currentHand)...)
		// possibilities := append(FindAllRuns(currentHand), FindAllGroupsWithJokers(currentHand)...)
		possibilities := GetAllPossibleCombos(currentHand)

		// Si aucune combinaison n'est possible, on compare avec notre meilleur score
		if len(possibilities) == 0 {
			if len(currentHand) < minTilesLeft {
				minTilesLeft = len(currentHand)
				bestTable = make([][]Tile, len(currentTable))
				copy(bestTable, currentTable)
				remainingHand = currentHand
			}
			return
		}

		// On essaie chaque possibilité
		for _, combo := range possibilities {
			newHand := removeTiles(currentHand, combo)
			newTable := append(currentTable, combo)
			backtrack(newHand, newTable)

			// Optimisation : si on a 0 tuile, on a trouvé une solution parfaite
			if minTilesLeft == 0 {
				return
			}
		}
	}

	backtrack(hand, [][]Tile{})
	return bestTable, remainingHand
}

// ----------------------------------------------------------------------------
// DrawTile()
// ----------------------------------------------------------------------------
func (state *GameState) DrawTile() {
	if len(state.Remaining) == 0 {
		fmt.Println("La pioche est vide !")
		return
	}

	// Prendre la première tuile de la pioche
	tile := state.Remaining[0]
	state.Remaining = state.Remaining[1:]

	// L'ajouter à la main du joueur actuel
	currentPlayer := &state.Players[state.CurrentPlayerID]
	currentPlayer.Hand = append(currentPlayer.Hand, tile)

	fmt.Printf("Joueur %d pioche une tuile.\n", currentPlayer.ID)
}

// ----------------------------------------------------------------------------
// PrintTable()
// ----------------------------------------------------------------------------
func PrintTable(table [][]Tile) {
	fmt.Println(strings.Repeat("═", 80))
	if len(table) == 0 {
		fmt.Println("          (La table est actuellement vide)")
	} else {
		for i, combo := range table {
			fmt.Printf(" [%2d] : ", i+1)
			for _, tile := range combo {
				fmt.Print(FormatTile(tile))
			}
			fmt.Println()
		}
	}
	fmt.Println(strings.Repeat("═", 80))
}

// ----------------------------------------------------------------------------
// FormatTile()
// ----------------------------------------------------------------------------
func FormatTile(tile Tile) string {
	if tile.Value == 0 {
		return "  😁 "
	}
	colorIcon := ""
	colorCode := ""
	switch tile.Color {
	case Red:
		colorIcon = "🔴"
		colorCode = ColorRed
	case Blue:
		colorIcon = "🔵"
		colorCode = ColorBlue
	case Green:
		colorIcon = "🟢"
		colorCode = ColorGreen
	case Orange:
		colorIcon = "🟠"
		colorCode = ColorOrange
	}
	return fmt.Sprintf("%s%2d%s%s ", colorCode, tile.Value, colorIcon, ColorReset)
}

// ----------------------------------------------------------------------------
// TotalValue()
// ----------------------------------------------------------------------------
func (combo Combination) TotalValue() int {
	sum := 0
	// Pour une version simple : on somme les valeurs réelles.
	// Note : Un Joker (0) devrait normalement prendre la valeur
	// de la tuile qu'il remplace dans la suite ou le groupe.
	for _, tile := range combo {
		sum += tile.Value
	}
	return sum
}

// ----------------------------------------------------------------------------
// HumanTurn()
// ----------------------------------------------------------------------------
func (state *GameState) HumanTurn(p *Player) {
	// 1. On crée une copie de sauvegarde (pour pouvoir annuler)
	backupTable := cloneTable(state.Table)
	backupHand := cloneHand(p.Hand)

	// 2. On crée le "Vrac" : une liste plate de toutes les tuiles disponibles
	pool := []Tile{} // On commence avec une réserve VIDE

	for {
		// On nettoie l'écran (optionnel mais recommandé)
		fmt.Print("\033[H\033[2J")

		currentValueOnTable := calculateTotalValue(state.Table)
		backupValueOnTable := calculateTotalValue(backupTable)
		pointsFromHand := currentValueOnTable - backupValueOnTable

		// On affiche le menu propre
		state.PrintUserMenu(p, pool, pointsFromHand)

		// 2. Lecture de l'action
		var action string
		fmt.Scanln(&action)
		action = strings.ToLower(action)

		// 3. Traitement des actions
		switch action {
		case "h":
			// On déplace une tuile de la main vers la réserve (pool)
			idx := getIndex()
			pool = append(pool, p.Hand[idx])
			p.Hand = append(p.Hand[:idx], p.Hand[idx+1:]...)

		case "n":
			// On choisit des tuiles dans la RÉSERVE pour former une suite/groupe
			indices := getMultipleIndices()
			var newCombo Combination
			for _, i := range indices {
				newCombo = append(newCombo, pool[i])
			}

			if IsValidCombination(newCombo) {
				SortTiles(newCombo)
				state.Table = append(state.Table, newCombo)
				pool = removeTilesFromPool(pool, indices)
				fmt.Println("✅ Nouvelle combinaison posée !")
			}

		case "p":
			// Annulation : on restaure l'état initial
			state.Table = backupTable
			p.Hand = backupHand

			// Puis on effectue la pioche obligatoire
			state.DrawTile()
			fmt.Println("Tour annulé, vous avez pioché une tuile.")
			return

		case "m":
			idx := getIndex()
			if idx < 0 || idx >= len(pool) {
				fmt.Println("❌ Index invalide.")
				continue
			}

			tuileAChecker := pool[idx]

			// Vérification : la tuile vient-elle de la main originale ?
			estOriginale := false
			for _, t := range backupHand {
				// On compare Valeur et Couleur (et éventuellement un ID unique si tu en as un)
				if t.Value == tuileAChecker.Value && t.Color == tuileAChecker.Color {
					estOriginale = true
					break
				}
			}

			if estOriginale {
				p.Hand = append(p.Hand, tuileAChecker)
				pool = removeTilesFromPool(pool, []int{idx})
				fmt.Println("✅ Tuile remise en main.")
			} else {
				fmt.Println("❌ Interdit ! Cette tuile provient de la table, vous ne pouvez pas la prendre dans votre main.")
			}

		case "t":
			if len(state.Table) == 0 {
				fmt.Println("❌ La table est vide.")
				continue
			}
			fmt.Printf("Quelle combinaison voulez-vous ramasser (1 à %d) ? ", len(state.Table))
			var idx int
			fmt.Scanln(&idx)
			idx -= 1 // Ajustement pour l'index 0

			if idx >= 0 && idx < len(state.Table) {
				// 1. On récupère les tuiles
				comboARamasser := state.Table[idx]

				// 2. On les ajoute au pool
				pool = append(pool, comboARamasser...)

				// 3. On les supprime de la table
				state.Table = append(state.Table[:idx], state.Table[idx+1:]...)

				fmt.Println("✅ Combinaison envoyée dans la réserve.")
			} else {
				fmt.Println("❌ Index invalide.")
			}

		case "s":
			// 1. Vérification des tuiles orphelines
			if len(pool) > 0 {
				fmt.Printf("❌ Action impossible : il reste %d tuile(s) dans la réserve !\n", len(pool))
				fmt.Println("Vous devez toutes les replacer sur la table.")
				continue // On retourne au début de la boucle pour que le joueur continue
			}

			// 1bis. Vérification : le joueur doit avoir posé au moins une tuile de sa main
			if len(p.Hand) == len(backupHand) {
				fmt.Println("❌ Vous n'avez posé aucune tuile de votre main. Tour annulé, vous piochez une tuile.")
				state.Table = backupTable
				p.Hand = backupHand
				state.DrawTile()
				return
			}

			// 2. Calcul du score pour la première pose (si nécessaire)
			if !p.HasPlayedFirst {
				if pointsFromHand < 30 {
					fmt.Printf("❌ Premier coup invalide : vous avez posé %d points (minimum 30).\n", pointsFromHand)
					fmt.Println("Souhaitez-vous [c]ontinuer ou [a]nnuler et piocher ?")

					var choice string
					fmt.Scanln(&choice)
					if choice == "a" {
						// On restaure tout et on pioche (Action "p")
						state.Table = backupTable
						p.Hand = backupHand
						state.DrawTile()
						return
					}
					continue // On revient au tour pour qu'il ajoute des tuiles
				}

				// Si on arrive ici, c'est que le score est >= 30
				p.HasPlayedFirst = true
				fmt.Println("🎉 Félicitations ! Vous avez validé votre ouverture.")
			}

			// 3. Validation finale
			fmt.Println("✅ Tour validé. Fin du tour.")
			return // On sort de HumanTurn, state.Table contient déjà les nouvelles modifs

		case "q":
			fmt.Print("Êtes-vous sûr de vouloir quitter ? (o/n) : ")
			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(confirm) == "o" {
				fmt.Println("Fermeture du jeu...")
				os.Exit(0)
			}
		}
	}
}

// ----------------------------------------------------------------------------
// DetermineFirstPlayer()
// ----------------------------------------------------------------------------
func (state *GameState) DetermineFirstPlayer() int {
	// Initialisation de la graine aléatoire (pour ne pas avoir toujours le même résultat)
	rand.Seed(time.Now().UnixNano())

	// On tire un nombre entre 0 et le nombre de joueurs - 1
	firstPlayerIndex := rand.Intn(len(state.Players))

	fmt.Printf("\n🎲 Tirage au sort... C'est %s qui commence !\n", state.Players[firstPlayerIndex].Name)
	// On laisse un petit temps de pause pour le suspense
	time.Sleep(2 * time.Second)

	return firstPlayerIndex
}

// ----------------------------------------------------------------------------
// calculateTotalValue()
// ----------------------------------------------------------------------------
func calculateTotalValue(table [][]Tile) int {
	total := 0
	for _, combo := range table {
		isRun := IsValidRun(combo)
		total += GetComboValueWithJoker(combo, isRun)
	}
	return total
}

// ----------------------------------------------------------------------------
// PrintHandWithIndices()
// ----------------------------------------------------------------------------
func (p Player) PrintHandWithIndices() {
	SortTiles(p.Hand)

	fmt.Printf("Votre main :\n")
	var lastColor Color = -2 // Valeur bidon pour le début

	for i, tile := range p.Hand {
		// Si la couleur change (et que ce n'est pas un Joker), on peut mettre un petit espace
		if i > 0 && tile.Color != lastColor && tile.Value != 0 {
			fmt.Print("\n")
		}
		fmt.Printf("[%2d]:%s ", i, FormatTile(tile))
		lastColor = tile.Color
	}
	fmt.Println()
}

// ----------------------------------------------------------------------------
// parseIndices()
// ----------------------------------------------------------------------------
func parseIndices(input string) []int {
	var indices []int
	parts := strings.Split(input, ",")
	for _, p := range parts {
		var idx int
		fmt.Sscanf(strings.TrimSpace(p), "%d", &idx)
		indices = append(indices, idx)
	}
	return indices
}

// ----------------------------------------------------------------------------
// IATurn()
// ----------------------------------------------------------------------------
func (state *GameState) IATurn(currentPlayer *Player) {
	fmt.Println("L'ordinateur réfléchit...")
	// currentPlayer := &state.Players[state.CurrentPlayerID]

	// 1. L'IA analyse sa main pour trouver la MEILLEURE configuration
	// (Cette fonction retourne toutes les combinaisons possibles et les tuiles restantes)
	bestLayout, remainingHand := FindBestHandLayout(currentPlayer.Hand)

	// 2. Calculer la valeur totale de ce qu'on s'apprête à poser
	totalProposed := 0
	for _, combo := range bestLayout {
		// On utilise notre logique de points (en considérant que ce sont des groupes/suites valides)
		// Pour simplifier ici, on utilise une détection rapide
		isRun := IsValidRun(combo)
		totalProposed += GetComboValueWithJoker(combo, isRun)
	}

	// 3. Logique de décision selon la règle des 30 points
	canPlay := false

	if !currentPlayer.HasPlayedFirst {
		if totalProposed >= 30 {
			canPlay = true
			currentPlayer.HasPlayedFirst = true
			fmt.Printf("⭐ Joueur %d : Première pose validée avec %d points !\n", currentPlayer.ID, totalProposed)
		}
	} else {
		// Si déjà ouvert, on joue dès qu'on peut poser au moins une tuile
		if len(remainingHand) < len(currentPlayer.Hand) {
			canPlay = true
		}
	}

	// 4. Action !
	if canPlay {
		// On ajoute les nouvelles combinaisons à la table
		state.Table = append(state.Table, bestLayout...)
		// On met à jour la main du joueur
		currentPlayer.Hand = remainingHand
		state.ConsecutivePasses = 0 // Quelqu'un a joué ! On remet le compteur à zéro.
	} else {
		// Le joueur ne peut pas jouer (ou n'a pas assez de points pour la 1ère fois)
		state.DrawTile()
		state.ConsecutivePasses++ // On incrémente car le joueur n'a rien pu poser
	}
}

// ----------------------------------------------------------------------------
// GetComboValueWithJoker()
// ----------------------------------------------------------------------------
func GetComboValueWithJoker(combo Combination, isRun bool) int {
	if len(combo) == 0 {
		return 0
	}

	// 1. Extraire les valeurs réelles et compter les Jokers
	var realValues []int
	jokers := 0
	for _, t := range combo {
		if t.Value == 0 {
			jokers++
		} else {
			realValues = append(realValues, t.Value)
		}
	}

	if len(realValues) == 0 {
		return 0
	}
	sort.Ints(realValues)

	if !isRun {
		// Groupe : Toutes les tuiles valent la valeur de la tuile réelle
		return realValues[0] * len(combo)
	}

	// Suite : Calculer la somme en tenant compte des trous comblés par les Jokers
	total := 0
	for _, v := range realValues {
		total += v
	}

	// On remplit les trous entre les tuiles réelles
	for i := 0; i < len(realValues)-1; i++ {
		for v := realValues[i] + 1; v < realValues[i+1]; v++ {
			total += v
			jokers--
		}
	}

	// On place les jokers restants aux extrémités (priorité au haut, max 13)
	low := realValues[0]
	high := realValues[len(realValues)-1]
	for jokers > 0 {
		if high < 13 {
			high++
			total += high
		} else {
			low--
			total += low
		}
		jokers--
	}
	return total
}

// ----------------------------------------------------------------------------
// FindAllGroupsWithJokers()
// ----------------------------------------------------------------------------
func FindAllGroupsWithJokers(hand []Tile) []Combination {
	var groups []Combination
	byValue := make(map[int]map[Color]Tile)
	jokersCount := 0

	// 1. On trie par valeur et on filtre les couleurs en même temps
	for _, t := range hand {
		if t.Value == 0 {
			jokersCount++
		} else {
			if byValue[t.Value] == nil {
				byValue[t.Value] = make(map[Color]Tile)
			}
			byValue[t.Value][t.Color] = t
		}
	}

	// 2. On analyse chaque groupe de valeur
	for _, colorMap := range byValue {
		// On transforme la map de couleurs en slice pour pouvoir manipuler les tuiles
		var tilesInValue []Tile
		for _, t := range colorMap {
			tilesInValue = append(tilesInValue, t)
		}

		count := len(tilesInValue)

		// Cas A : 3 ou 4 tuiles de couleurs différentes
		if count >= 3 {
			groups = append(groups, Combination(tilesInValue))
		}

		// Cas B : 2 ou 3 tuiles réelles + 1 Joker (pour faire un groupe de 3 ou 4)
		if (count == 2 || count == 3) && jokersCount >= 1 {
			combo := append(Combination{}, tilesInValue...)
			combo = append(combo, Tile{Value: 0, Color: -1}) // Ajout du Joker
			groups = append(groups, combo)
		}

		// Note : On pourrait aussi gérer 1 tuile + 2 Jokers,
		// mais c'est souvent moins stratégique pour l'IA au début.
	}
	return groups
}

// ----------------------------------------------------------------------------
// FindAllRunsWithJokers()
// ----------------------------------------------------------------------------
func FindAllRunsWithJokers(hand []Tile) []Combination {
	var allRuns []Combination

	// 1. Séparer par couleur et compter les jokers
	byColor := make(map[Color][]Tile)
	jokersInHand := 0
	for _, t := range hand {
		if t.Value == 0 {
			jokersInHand++
		} else {
			byColor[t.Color] = append(byColor[t.Color], t)
		}
	}

	// 2. Pour chaque couleur, on cherche des suites
	for _, tiles := range byColor {
		// On trie et on enlève les doublons pour faciliter la recherche de suites
		sortedTiles := uniqueSortedTiles(tiles)

		// On teste toutes les fenêtres de départ possibles (valeur de 1 à 13)
		for startVal := 1; startVal <= 11; startVal++ {
			// Et toutes les longueurs possibles (3 à 13)
			for length := 3; length <= 13; length++ {
				if startVal+length-1 > 13 {
					break
				}

				currentCombo := Combination{}
				jokersNeeded := 0

				// On essaie de construire la suite [startVal ... startVal+length-1]
				for v := startVal; v < startVal+length; v++ {
					found := false
					for _, t := range sortedTiles {
						if t.Value == v {
							currentCombo = append(currentCombo, t)
							found = true
							break
						}
					}
					if !found {
						// On n'a pas la tuile, on aura besoin d'un joker
						currentCombo = append(currentCombo, Tile{Value: 0, Color: -1})
						jokersNeeded++
					}
				}

				// Si on a assez de jokers et qu'on n'a pas fait une suite "que de jokers"
				if jokersNeeded <= jokersInHand && jokersNeeded < length {
					// On fait une copie propre pour l'ajouter aux résultats
					runCopy := make(Combination, len(currentCombo))
					copy(runCopy, currentCombo)
					allRuns = append(allRuns, runCopy)
				}
			}
		}
	}
	return allRuns
}

// ----------------------------------------------------------------------------
// uniqueSortedTiles()
// ----------------------------------------------------------------------------
// Fonction utilitaire pour trier et ignorer les doublons (ex: deux 7 rouges)
func uniqueSortedTiles(tiles []Tile) []Tile {
	if len(tiles) == 0 {
		return tiles
	}
	sort.Slice(tiles, func(i, j int) bool { return tiles[i].Value < tiles[j].Value })
	unique := []Tile{tiles[0]}
	for i := 1; i < len(tiles); i++ {
		if tiles[i].Value != tiles[i-1].Value {
			unique = append(unique, tiles[i])
		}
	}
	return unique
}

// ----------------------------------------------------------------------------
// GetAllPossibleCombos()
// ----------------------------------------------------------------------------
func GetAllPossibleCombos(hand []Tile) []Combination {
	var all []Combination
	all = append(all, FindAllGroupsWithJokers(hand)...)
	all = append(all, FindAllRunsWithJokers(hand)...)
	return all
}

// ----------------------------------------------------------------------------
// cloneHand()
// ----------------------------------------------------------------------------
func cloneHand(hand []Tile) []Tile {
	newHand := make([]Tile, len(hand))
	copy(newHand, hand)
	return newHand
}

// ----------------------------------------------------------------------------
// cloneTable()
// ----------------------------------------------------------------------------
func cloneTable(table [][]Tile) [][]Tile {
	newTable := make([][]Tile, len(table))
	for i, combo := range table {
		newTable[i] = make([]Tile, len(combo))
		copy(newTable[i], combo)
	}
	return newTable
}

// ----------------------------------------------------------------------------
// PrintTilePool()
// ----------------------------------------------------------------------------
func PrintTilePool(pool []Tile) {
	if len(pool) == 0 {
		fmt.Println("    [ Vide ]")
		return
	}
	fmt.Print("RÉSERVE : ")
	for i, t := range pool {
		fmt.Printf("[%2d]:%s ", i, FormatTile(t))
	}
	fmt.Println()
}

// ----------------------------------------------------------------------------
// getIndex()
// ----------------------------------------------------------------------------
func getIndex() int {
	var idx int
	fmt.Print("Entrez l'index : ")
	fmt.Scanln(&idx)
	return idx
}

// ----------------------------------------------------------------------------
// getMultipleIndices()
// ----------------------------------------------------------------------------
func getMultipleIndices() []int {
	fmt.Print("Entrez les index séparés par des virgules (ex: 0,1,2) : ")
	var input string
	fmt.Scanln(&input)

	// On réutilise ta logique parseIndices
	return parseIndices(input)
}

// ----------------------------------------------------------------------------
// removeTilesFromPool()
// ----------------------------------------------------------------------------
func removeTilesFromPool(pool []Tile, indices []int) []Tile {
	// 1. Trier les index du plus grand au plus petit
	sort.Sort(sort.Reverse(sort.IntSlice(indices)))

	for _, idx := range indices {
		if idx >= 0 && idx < len(pool) {
			// Technique standard en Go pour supprimer un élément d'un slice
			pool = append(pool[:idx], pool[idx+1:]...)
		}
	}
	return pool
}

// ----------------------------------------------------------------------------
// PrintUserMenu()
// ----------------------------------------------------------------------------
func (state *GameState) PrintUserMenu(p *Player, pool []Tile, points int) {
	// Calcul du nombre de tuiles restantes
	remainingTiles := len(state.Remaining)

	fmt.Println("\n" + strings.Repeat("═", 80))
	fmt.Println("                                                        _ ")
	fmt.Println("                    __ _ _ __ _   _ _ __ ___  _ __ ___ (_)")
	fmt.Println("                   / _` | '__| | | | '_ ` _ \\| '_ ` _ \\| |")
	fmt.Println("                  | (_| | |  | |_| | | | | | | | | | | | |")
	fmt.Println("                   \\__, |_|   \\__,_|_| |_| |_|_| |_| |_|_|")
	fmt.Println("                   |___/                        © JPL 2026")
	fmt.Println("\n" + strings.Repeat("═", 80))
	// Ajout de la pioche dans le bandeau
	fmt.Printf(" 👤 JOUEUR : %-12s | 🃏 PIOCHE : %-3d | 🏆 OUVERTURE : %s\n",
		p.Name, remainingTiles, formatStatus(p.HasPlayedFirst))
	fmt.Print(" 👥 TUILES : ")
	for _, other := range state.Players {
		if other.HasPlayedFirst {
			fmt.Printf("%s[✔ %-7s: %2d]%s  ", ColorGreen, other.Name, len(other.Hand), ColorReset)
		} else {
			fmt.Printf("%s[✖ %-7s: %2d]%s  ", ColorRed, other.Name, len(other.Hand), ColorReset)
		}
	}
	fmt.Println() // Pour finir la ligne proprement
	fmt.Println("\n" + strings.Repeat("─", 80))

	// 1. État de la Table (Ce qui est validé)
	fmt.Println(" 🧩 TABLE ACTUELLE :")
	if len(state.Table) == 0 {
		fmt.Println("    [ Vide ]")
	} else {
		PrintTable(state.Table)
	}

	// 2. La Réserve (Le vrac à traiter)
	fmt.Println("\n 📥 RÉSERVE (Tuiles à replacer) :")
	PrintTilePool(pool)

	// 3. La Main du joueur
	fmt.Println("\n 🖐️  VOTRE MAIN :")
	p.PrintHandWithIndices()

	// 4. Menu des actions
	fmt.Println("\n" + strings.Repeat("─", 80))
	fmt.Printf(" ✨ Points de ce tour : %d / 30\n", points)
	fmt.Println("  [N] Nouveau combo        (Réserve -> Table)")
	fmt.Println("  [H] Jouer de la main     (Main    -> Réserve)")
	fmt.Println("  [T] Ramasser de la table (Table   -> Réserve)")
	fmt.Println("  [M] Reprendre en main    (Réserve -> Main)")
	fmt.Println("  [S] Valider le tour")
	fmt.Println("  [P] Piocher et/ou annuler le tour")
	fmt.Println("  [Q] Quitter le jeu")
	fmt.Println(strings.Repeat("═", 80))
	fmt.Print("👉 Votre choix : ")
}

// ----------------------------------------------------------------------------
// formatStatus()
// ----------------------------------------------------------------------------
func formatStatus(b bool) string {
	if b {
		return "✅ FAITE"
	}
	return "❌ À FAIRE"
}
