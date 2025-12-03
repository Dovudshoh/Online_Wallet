package user

import "time"

type User struct {
	ID        int       
	Name      string    
	Email     string    
	Password  string    
	BalanceTJS float64  
	BalanceUSD float64  
	BalanceEUR float64  
	CreatedAt time.Time 
}
