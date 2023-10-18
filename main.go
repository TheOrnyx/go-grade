// TODO
// Possible like have user input what score was out of etc
// Convert to an actual TUI interface

package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/fatih/color"
	_ "github.com/mattn/go-sqlite3"
)

const (
	ADD_CLASS = iota
	ADD_ASSIGNMENT
	ADD_GRADE

	LIST_CLASSES
	EXIT
)

func main() {
	gradeDB, _ := openDBFile("grades.db")
	defer gradeDB.Close()
	createGradeTables(gradeDB)
	running := true

	for running {
		switch getUserOption() {
		case ADD_CLASS :
			addNewClass(gradeDB)

		case LIST_CLASSES :
			listClasses(gradeDB)

		case ADD_GRADE :
			addGradeMenu(gradeDB)
			
		case EXIT :
			running = false
			color.Red("Exiting")
		}
	}
}

func getUserOption() int {
	var choice int
	fmt.Println("Please choose an option")
	fmt.Println("| 0: Add A Class")
	fmt.Println("| 1: Add An Assignment")
	fmt.Println("| 2: Add A Grade")
	fmt.Println("| 3: List Classes")
	fmt.Println("| 4: EXIT")
	fmt.Println("|--------------------")
	fmt.Printf("> ")
	fmt.Scanln(&choice)
	//fmt.Println("---------------------")
	return choice
}

func listClasses(db *sql.DB) {
	fmt.Println("---------------------")
	rows := getClassRows(db)

	for rows.Next() {
		var className string
		var classID int
		rows.Scan(&classID,&className)
		fmt.Println("├──",className)
		listClassAssignments(db, classID)
		fmt.Println("│   ")
		fmt.Println("│   ")
	}
	fmt.Println("---------------------")
	color.New(color.FgHiBlack).Printf("Press enter to return")
	fmt.Scanln()
}

func listClassAssignments(db *sql.DB, classID int){
	rows := getClassAssignments(db, classID)

	for rows.Next() {
		var id int
		var assName string
		var weight, grade float64
		rows.Scan(&id, &assName, &weight, &grade)
		fmt.Println("│   ├──", assName)
		fmt.Println("│   │   ├── Ass Weight: ", weight)
		fmt.Println("│   │   ├── Ass Grade: ", grade,"%")
	}

}

func getClassRows(db *sql.DB) *sql.Rows {
	rows, err := db.Query("SELECT ClassID, ClassName FROM Classes;")
	checkErr(err)
	return rows
}

func getClassAssignments(db *sql.DB, classID int) *sql.Rows {
	rows, err := db.Query("SELECT AssignmentID, AssignmentName, Weight, Grade FROM Assignments WHERE ClassID=?", classID)
	checkErr(err)
	return rows
}



func addNewClass(db *sql.DB) {
	var className string
	var assNum, classID int
	fmt.Println("Please enter the class name")
	fmt.Printf("> ")
	fmt.Scanln(&className)
	classID = insertNewClass(db, className)

	fmt.Println("\nHow many assignments does this class have?")
	fmt.Printf("> ")
	fmt.Scanln(&assNum)

	for i := 0; i < assNum; i++ {
		var assName string
		var assWeight float64
		fmt.Println("----")
		fmt.Printf("Assignment Name: > ")
		scan := bufio.NewScanner(os.Stdin)
		if scan.Scan() {
			assName = scan.Text()
		} else {
			color.Red("Error reading input", scan.Err())
		}
		fmt.Printf("Assignment Weight: > ")
		fmt.Scanln(&assWeight)
		insertNewAssignment(db, classID, assName, assWeight)
	}
}

func addGradeMenu(db *sql.DB) {
	var classID, assignmentID, lastID int
	lastID = 1
	fmt.Println("---------------------")
	fmt.Println("| Choose a class ID from Below:")
	rows := getClassRows(db)
	
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		fmt.Printf("| %d  %s\n", id, name)
	}
	fmt.Printf("\n> ")
	fmt.Scanln(&classID)

	for assignmentID != lastID {
		fmt.Println("---------------------")
		assignmentID, lastID = chooseAssignment(db, classID)
		if assignmentID != lastID {
			changeGrade(db, assignmentID)
		}
	}
}

func changeGrade(db *sql.DB, gradeID int) {
	var maxScore, userScore int
	var scorePercent float64
	fmt.Println("Beginning grade change")
	fmt.Printf("Max Score: > ")
	fmt.Scanln(&maxScore)
	fmt.Printf("Your Score: > ")
	fmt.Scanln(&userScore)
	scorePercent = float64(userScore) / float64(maxScore) * 100 //maybe add a 0 check
	scorePercent = math.Floor(scorePercent * 100) / 100
	fmt.Println("Your score was", scorePercent, "%")
	fmt.Println("Updating score...")
	updateGrade(db, gradeID, scorePercent)
}

func updateGrade(db *sql.DB, gradeID int, gradeScore float64) {
	query, err := db.Prepare("UPDATE Assignments SET Grade=? WHERE AssignmentID=?")
	checkErr(err)
	defer query.Close()

	_, err = query.Exec(gradeScore, gradeID)
	checkErr(err)
	c:= color.New(color.FgHiBlack)
	c.Println("Grade updated")
	fmt.Scanln()
}

func chooseAssignment(db *sql.DB, classID int) (int, int){
	var assignmentID, lastID int
	fmt.Println("| Choose an Assignment ID from below:")
	rows := getClassAssignments(db, classID)
	
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name, nil, nil)
		fmt.Printf("| %d  %s\n", id, name)
		lastID = id
	}
	color.Red("| %d Return", lastID + 1)
	
	fmt.Printf("\n> ")
	fmt.Scanln(&assignmentID)
	fmt.Println("---------------------")
	return assignmentID, lastID + 1
}

func insertNewAssignment(db *sql.DB, classID int, assName string, assWeight float64) {
	query, err := db.Prepare("INSERT INTO Assignments(ClassID, AssignmentName, Weight, Grade) VALUES(?, ?, ?, 0)")
	if err != nil {
		log.Fatal(err)
	}
	defer query.Close()

	_, err = query.Exec(classID, assName, assWeight)
	if err != nil {
		log.Fatal(err)
	}
	c := color.New(color.FgHiBlack)
	c.Println("Assignment added")
}

func insertNewClass(db *sql.DB, className string) int {
	var classID int
	_, err := db.Exec("INSERT OR IGNORE INTO Classes(ClassName) VALUES(?)", className)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Class " + className + " inserted")
	}

	row := db.QueryRow("SELECT ClassID FROM Classes WHERE ClassName=?", className)
	err = row.Scan(&classID)
	if err != nil {
		log.Fatal(err)
	}

	return classID
}

func openDBFile(DBLoc string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", DBLoc)
	if err != nil {
		log.Fatal(err)
	}
	
	return db, err
}

func createGradeTables(db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS Classes (ClassID INTEGER PRIMARY KEY, ClassName TEXT, UNIQUE(ClassName));")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Assignments (AssignmentID INTEGER PRIMARY
                    KEY, ClassID INTEGER, AssignmentName TEXT, Weight REAL, Grade REAL, FOREIGN KEY
                    (ClassID) REFERENCES Classes(ClassID));`)
	if err != nil {
		log.Fatal(err)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
