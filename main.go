/**
Yarrrrrrrrrrrr, here be code. Thanks Chris, ye may now walk the plank. :D

The Go Muskateers

By 	Hannah McKenna - 17204178
	Robert O'Shea - 13147021
	Ross Duffy - 17201624
	Thomas Langley - 17215145
*/

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Sets some constants that will be used throughout the program
const (
	maxCustCount                    = 75
	maxTillQueueLength              = 6
	maxItemsToBuy                   = 200
	maxPotentialCustomerArrayLength = 100
	maxTillOperatorCount            = 6
	amountOfItemLimitTills          = 2
)

// Sets some variables that will be used throughout the program
var (
	weather                                     = setWeather()
	fullCustomerList                            = getListOfCustomers(maxPotentialCustomerArrayLength)
	customerListComfortableWithBadWeather       = getCustomersComfortableWithWeather(fullCustomerList, weather)
	theshop                                     = newShop()
	totalComfortableWithWeather                 = 0
	totalWaitInShopTime                   int64 = 0
	totalWaitInQueueTime                  int64 = 0
	totalUtilisation                      int64 = 0
)

/*
	Runs simulation:
	- Builds shop
	- Fills shop with tills
	- Fills shop with customers
	- Carries out sim
	- Prints report
*/
func main() {
	rand.Seed(time.Now().UnixNano())
	custsChan := make(chan customer)
	wg := sync.WaitGroup{}

	go produce(custsChan, customerListComfortableWithBadWeather)

	wg.Add(len(customerListComfortableWithBadWeather))

	for i := 0; i < maxTillOperatorCount; i++ {
		go processQueue(i)
	}

	totalComfortableWithWeather = len(customerListComfortableWithBadWeather)
	fmt.Println(totalComfortableWithWeather)
	go accommodateCustomers(custsChan, len(customerListComfortableWithBadWeather), &wg)

	wg.Wait()
	totalProcessed := 0
	for i := 0; i < maxTillOperatorCount; i++ {
		for len(theshop.tills[i].customerQueue) > 0 {
			time.Sleep(1 * time.Second)
		}
		totalProcessed += theshop.tills[i].itemsProcessed
	}

	for i := 0; i < maxTillOperatorCount; i++ {
		fmt.Println("Till ", i, " Total Customers: ", theshop.tills[i].customersProcessed)
	}
	fmt.Println("Till Count: ", len(theshop.tills))
	fmt.Println("Average Till Utilisation:         ", totalUtilisation/int64(maxTillOperatorCount))
	fmt.Println("Average items processed per till: ", totalProcessed/maxTillOperatorCount)
	fmt.Println("Total items processed:            ", totalProcessed)
	fmt.Println("Average customer wait time:       ", totalWaitInQueueTime/int64(totalUtilisation))
	fmt.Println("Average customer trolly contents: ", totalProcessed/(theshop.totCustEntered-theshop.custLost))
	fmt.Println("Lost customers:                   ", theshop.custLost)
}

/*
	Randomly moves customers down the channel
*/
func produce(jobs chan customer, custSlice []customer) {
	for _, cust := range custSlice {
		jobs <- cust
		time.Sleep(time.Duration(rand.Int31n(100)) * time.Millisecond)
	}
}

/*
	Adds customers who are ok with the weather to shop at random intervals
*/
func accommodateCustomers(customerIn chan customer, totalComing int, wg *sync.WaitGroup) {
	rand.Seed(time.Now().UnixNano())
	wg.Add(1)
	customersBrowsingWG := sync.WaitGroup{}
	for i := 0; i < totalComfortableWithWeather; i++ {
		c := <-customerIn
		if theshop.addCustToShop(&c) {
			customersBrowsingWG.Add(1)
			go theshop.handleBrowsing(c, &customersBrowsingWG)
		}
		wg.Done()
	}
	fmt.Println("Waiting for customers to finish browsing")
	customersBrowsingWG.Wait()
	wg.Done()
}

/*
	Infinite loop that processes the customers in the till's queues
*/
func processQueue(idx int) {
	for {
		if len(theshop.tills[idx].customerQueue) > 0 {
			theshop.tills[idx].customersProcessed++
			totalUtilisation++
			c := theshop.tills[idx].customerQueue[0]
			for i := 0; i < c.itemsToBuy; i++ {
				time.Sleep(10 * time.Millisecond)
				theshop.tills[idx].itemsProcessed++
			}
			timeLeftQueueAndShop := time.Now()
			t1 := timeLeftQueueAndShop.Sub(c.timeEnteredQueue)
			t2 := timeLeftQueueAndShop.Sub(c.timeEnteredShop)
			totalWaitInQueueTime += t1.Milliseconds()
			totalWaitInShopTime += t2.Milliseconds()
			fmt.Println("customer's time in queue ", t1)
			fmt.Println("customer's time in store ", t2)
			theshop.tills[idx].removeCustFromQueue(theshop.tills[idx].customerQueue[0])
			theshop.customerLeaves(c)
			theshop.custLost--
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

/*
	Removes customers from store after they are finished at the till
*/
func (S *shop) removeCustomFromStore(cust customer) {
	for i := 0; i < len(S.custInStore); i++ {
		if (S.custInStore[i] == cust) && (len(S.custInStore) >= 1) {
			S.custInStore = append(S.custInStore[:i], S.custInStore[i+1:]...)
			break
		}
	}
}

/*
	Checks for customers who are OK with the weather
*/
func getCustomersComfortableWithWeather(fullCustomerList []customer, weatherValue bool) []customer {
	comfortableList := []customer{}

	for i := 0; i < maxPotentialCustomerArrayLength; i++ {
		if !weatherValue {
			if fullCustomerList[i].comfortableWithBadWeather {
				comfortableList = append(comfortableList, fullCustomerList[i])
			}
		} else {
			comfortableList = append(comfortableList, fullCustomerList[i])
		}
	}

	return comfortableList
}

//------------------------------//
//	 This defines the weather	//
//------------------------------//

/*
	Sets weather values
*/
func setWeather() bool {

	weatherNumber := rand.Intn(11)
	var weatherType string

	if weatherNumber <= 4 {
		weatherType = "sunny"
	} else if weatherNumber == 9 {
		weatherType = "snowy"
	} else if weatherNumber == 10 {
		weatherType = "stormy"
	} else {
		weatherType = "rainy"
	}

	fmt.Println("The weather today is", weatherType)

	return weatherNumber <= 4
}

//--------------------------//
//	This defines the shop	//
//--------------------------//

/*
	Definition of shop struct
*/
type shop struct {
	//item stock
	stock               int
	totalProcessedItems int

	custInStore              []customer
	totCustEntered, custLost int

	tills []tillOperator

	addCustomerMutex sync.Mutex
}

/*
	Method to create new shop
*/
func newShop() shop {
	tills := make([]tillOperator, 0, maxTillOperatorCount)

	for i := 0; i < maxTillOperatorCount; i++ {
		if i > amountOfItemLimitTills {
			tills = append(tills, newTillOperatorWithoutItemLimit())
		} else {
			tills = append(tills, newTillOperatorWithItemLimit())
		}
	}
	return shop{
		getStockCount(),                   // stockCount
		0,                                 // totalProcessedItems
		make([]customer, 0, maxCustCount), // custInStore
		0,                                 // totCustEntered
		0,                                 // custBrowsing
		tills,                             // tills
		sync.Mutex{},
	}
}

/*
	Used to create new shop
*/
func getStockCount() int {
	return 10000
}

/*
	Method to pass a list of customers to a shop
*/
func (S *shop) fillCustomers(customers []customer) {
	S.totCustEntered = len(customers)
	S.custInStore = customers
}

/*
	Adds customer to queue,
	if customer has few enough items they can go to item limit tills
*/
func (S *shop) addCustToQueue(cust customer) *tillOperator {
	for i := 0; i < len(S.tills); i++ {
		if S.tills[i].getQueueLength() < maxTillQueueLength {
			if S.tills[i].getItemLimitBool() && (cust.getTrollyItems() <= S.tills[i].getItemLimitCount()) {
				S.tills[i].addCustToQueue(cust)
			} else if !S.tills[i].getItemLimitBool() {
				S.tills[i].addCustToQueue(cust)
			}
			return &S.tills[i]
		} else if i == len(S.tills)-1 {
			S.customerLeaves(cust)
		}
	}
	return nil

}

/*
	Add customers to list of customers in shop
*/
func (S *shop) addCustToShop(cust *customer) bool {
	if len(S.custInStore) < maxCustCount {
		cust.timeEnteredShop = time.Now()
		S.totCustEntered++
		S.custInStore = append(S.custInStore, *cust)
		return true
	}
	S.custLost++
	return false
}

/*
	Customer leaves shop and is lost (without being served)
*/
func (S *shop) customerLeaves(cust customer) {
	for i := 0; i < len(S.custInStore); i++ {
		if S.custInStore[i].id == cust.id {
			theshop.custLost++
			S.custInStore = append(S.custInStore[:i], S.custInStore[i+1:]...)
			break
		}
	}
}

/*
	Counts total items processed by all tills
*/
func (S *shop) addToTotal(n int) {
	for i := 0; i < len(S.tills); i++ {
		S.totalProcessedItems += S.tills[i].itemsProcessed
	}
}

/*
	Method to pass a list of tills to shop
*/
func (S *shop) fillTills(tills []tillOperator) {
	S.tills = tills
}

/*
	Customers browses shop and records time customer entered
*/
func (S *shop) handleBrowsing(cust customer, wg *sync.WaitGroup) {
	cust.doShop()
	S.addCustomerMutex.Lock()
	cust.timeEnteredQueue = time.Now()
	S.addCustToQueue(cust)
	S.addCustomerMutex.Unlock()
	wg.Done()
}

//--------------------------//
//	This defines the Till	//
//--------------------------//

/*
	This defines the shop struct.
*/
type tillOperator struct {
	// Mandatory
	scanSpeedSeconds float64 // the speed at which they scan items, from 0.5s to 6s.

	// Mandatory, based on till.
	inUse          bool       // boolean to see if the till is open and items are being scanned.
	itemLimitBool  bool       // boolean to see if there is a limit on the number of items that a customer can have in their trolley to be able to pass through the till.
	itemLimitCount int        //  the max (or even min???) items per customer that the till is able to process.
	utilisation    float64    // the utilisation rate of the checkout. this contributes to the averageCheckoutUtilisation in shop.
	customerQueue  []customer // the list of customers queueing up at the till.

	attitude          int // the attitude the operator has towards the job. could determine the speed the operator scans items at. this is at a range of 0 to 100.
	temperament       int // the level of the temperament of the operator. the higher a temperament, the more likely there are to cause disagreements. this is at a range of 0 to 100.
	timeUntilShiftEnd int // the time until the end of the operator's shift. the closer they are to the end of their shift, the faster they may scan items. this could be a range of 0 to 8 (for hours?).

	itemsProcessed     int
	customersProcessed int
}

/*
	Creates a new till operator, with a specified item amount
*/
func newTillOperator(itemLimitCountAmount int) tillOperator {
	till := tillOperator{
		scanSpeedSeconds: getRandomOperatorScanSpeed(0.5, 6),

		inUse: false,

		itemLimitBool: checkItemLimitCountAmount(itemLimitCountAmount),

		itemLimitCount: itemLimitCountAmount,

		utilisation: 0,

		customerQueue: make([]customer, 0, 5),

		attitude: getRandomAttitudeValue(0, 100),

		temperament: getRandomTemperamentValue(0, 100),

		timeUntilShiftEnd: getRandomTimeUntilShiftEnd(0, 8),

		itemsProcessed: 0,

		customersProcessed: 0,
	}

	return till
}

/*
	Create a new till operator, without an item limit.
	Calls newTillOperator.
*/
func newTillOperatorWithoutItemLimit() tillOperator {
	till := newTillOperator(0)

	return till
}

/*
	Creates a new till operator, with an item limit of 5.
	Calls newTilLOperator.
*/
func newTillOperatorWithItemLimit() tillOperator {
	till := newTillOperator(5)

	return till
}

/*
	Gets if the till has an item limit.
*/
func (T tillOperator) getItemLimitBool() bool {
	return T.itemLimitBool
}

/*
	Gets the max item limits. 0 if there are no max items.
*/
func (T tillOperator) getItemLimitCount() int {
	return T.itemLimitCount
}

/*
	Determines if there is a max item limit count.
*/
func checkItemLimitCountAmount(itemLimitCountAmount int) bool {
	return itemLimitCountAmount > 0
}

/*
	Gets the random scan speed of the till's operator.
*/
func getRandomOperatorScanSpeed(min float64, max float64) float64 {
	randomScanSpeed := (rand.Float64()*(max-min) + min)
	return randomScanSpeed
}

/*
	Gets a random int value, with min and max values specified.
*/
func getRandomIntValue(min int, max int) int {
	randomValue := rand.Intn((max - min) + min)

	return randomValue
}

/*
	Gets a random attitude, with min and max values specified.
*/
func getRandomAttitudeValue(min int, max int) int {
	return getRandomIntValue(min, max)
}

/*
	Gets a random temperament value, with min and max values specified.
*/
func getRandomTemperamentValue(min int, max int) int {
	return getRandomIntValue(min, max)
}

/*
	Gets the time until shift end for the till, with min and max values specified.
*/
func getRandomTimeUntilShiftEnd(min int, max int) int {
	return getRandomIntValue(min, max)
}

/*
	Gets the till's queue length.
*/
func (T tillOperator) getQueueLength() int {
	return len(T.customerQueue)
}

/*
	Adds a customer to the till's queue.
*/
func (T *tillOperator) addCustToQueue(cust customer) {
	if T.getQueueLength() <= maxTillQueueLength {
		T.customerQueue = append(T.customerQueue, cust)
	} else {
		fmt.Println("Customer can't be added to queue")
		// customer leaves shop.
	}
}

/*
	Removes a customer from the till's queue.
*/
func (T *tillOperator) removeCustFromQueue(cust customer) {
	for i := 0; i < len(T.customerQueue); i++ {
		if T.customerQueue[i] == cust {
			T.customerQueue = append(T.customerQueue[:i], T.customerQueue[i+1:]...)
			break
		}
	}
}

/*
	Increases the count of items porcessed in the till.
*/
func (T *tillOperator) addToItemsProcess(newlyProcessedItems int) {
	T.itemsProcessed += newlyProcessedItems
}

//------------------------------//
// 	 This defines the customer	//
//------------------------------//

/*
	Used to determine how a customer is feelin i.e hungry
*/
type feeling int

/*
	Defines the customer struct
*/
type customer struct {
	id                        int
	comfortableWithBadWeather bool
	itemsToBuy, trolly        int     // number of items the customer wants to buy
	speed                     float64 // how quickly they can collect items
	status                    feeling // described in contants above
	timeEnteredShop           time.Time
	timeEnteredQueue          time.Time
	timeLeftQueue             time.Time
	timeLeftShop              time.Time
}

/*
	Generates a customer with random attributes
*/
func newCustomer(id int) customer {
	rand.Seed(time.Now().UnixNano())
	return customer{
		id:                        id,
		comfortableWithBadWeather: rand.Intn(10) > 4,
		itemsToBuy:                rand.Intn(maxItemsToBuy) + 1, // + 1 to avoid 0 items
		speed:                     rand.Float64(),
	}
}

/*
	This determines if a customer is comfortable with bad weather, i.e snow, rain or stormy weather
*/
func (c customer) isComfortableWithBadWeather() bool {
	return c.comfortableWithBadWeather
}

/*
	This gets the customer to pick up an item, which decrements the number of items they have left to pickup.
*/
func (c *customer) pickUpItem() bool {
	if c.itemsToBuy > 0 {
		c.itemsToBuy--
		return true
	}
	return false
}

/*
	Returns number of products that the customer has picked up
*/
func (c customer) getTrollyItems() int {
	return c.trolly
}

/*
	Prints out the relevant customer information.
*/
func (c customer) print() {
	fmt.Println("customer")

	if c.comfortableWithBadWeather {
		fmt.Println("Is happy with bad weather")
	} else {
		fmt.Println("Is not happy with bad weather")
	}

	fmt.Printf("Must buy %d items\n", c.itemsToBuy)

	fmt.Printf("Moves at %.2f speed\n", c.speed)
}

/*
	This has the customer "do their shopping", which would take a random number of milliseconds, from 0 to 100.
*/
func (c customer) doShop() {
	for i := 0; i < c.itemsToBuy; i++ {
		time.Sleep(100 * time.Millisecond)
	}
}

/*
	Creates a list of customers with a specified size.
*/
func getListOfCustomers(customerCount int) []customer {
	var customerList = make([]customer, customerCount)
	for i := 0; i < customerCount; i++ {
		customerList[i] = newCustomer(i)
	}
	return customerList
}
