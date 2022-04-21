package main

import (
	"context"
	"crypto/ecdsa"
	"famesensor/go-smart-contract/api"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

type DepositeBody struct {
	PrivateKey string `json:"privateKey"`
	Amount     int    `json:"amount"`
}

type Withdrawl = DepositeBody

func main() {
	// load env
	godotenv.Load()

	// address of etherum env
	client, err := ethclient.Dial("http://127.0.0.1:" + os.Getenv("ETHERUM_PORT"))
	if err != nil {
		panic(err)
	}

	auth := getAccountAuth(client, os.Getenv("PRIVATE_KEY"))

	//deploying smart contract
	address, tx, instance, err := api.DeployApi(auth, client)
	if err != nil {
		panic(err)
	}

	fmt.Println(address.Hex())

	fmt.Println("instance : ", instance)
	fmt.Println("tx : ", tx.Hash().Hex())

	conn, err := api.NewApi(common.HexToAddress(address.Hex()), client)
	if err != nil {
		panic(err)
	}

	app := fiber.New()
	app.Use(logger.New())

	app.Get("/balance", func(c *fiber.Ctx) error {
		reply, err := conn.Balance(&bind.CallOpts{})
		if err != nil {
			return err
		}
		return c.Status(fiber.StatusOK).JSON(reply)
	})

	app.Get("/admin", func(c *fiber.Ctx) error {
		reply, err := conn.Admin(&bind.CallOpts{})
		if err != nil {
			return err
		}
		return c.Status(fiber.StatusOK).JSON(reply)
	})

	app.Post("/deposite", func(c *fiber.Ctx) error {
		body := new(DepositeBody)

		if err := c.BodyParser(body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err})
		}

		auth := getAccountAuth(client, body.PrivateKey)

		reply, err := conn.Deposite(auth, big.NewInt(int64(body.Amount)))
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusOK).JSON(reply)
	})

	app.Post("/withdrawl", func(c *fiber.Ctx) error {
		body := new(Withdrawl)

		if err := c.BodyParser(body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err})
		}

		auth := getAccountAuth(client, body.PrivateKey)

		reply, err := conn.Withdrawl(auth, big.NewInt(int64(body.Amount)))
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusOK).JSON(reply)
	})

	if err := app.Listen(":" + os.Getenv("APP_PORT")); err != nil {
		panic(err)
	}
}

//function to create auth for any account from its private key
func getAccountAuth(client *ethclient.Client, privateKeyAddress string) *bind.TransactOpts {

	privateKey, err := crypto.HexToECDSA(privateKeyAddress)
	if err != nil {
		panic(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		panic("invalid key")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		panic(err)
	}
	fmt.Println("nounce=", nonce)
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		panic(err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)      // in wei
	auth.GasLimit = uint64(3000000) // in units
	auth.GasPrice = big.NewInt(1000000)

	return auth
}
