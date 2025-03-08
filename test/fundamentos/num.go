package fundamentos

import "fmt"

func num() {
	for i := 0; i <= 10; i++ {
		if i%2 == 0 {
			fmt.Println("\nNúmero par: ", i)
		} else {
			fmt.Print("\nNúmero ímpar: ", i)
		}
	}
}
