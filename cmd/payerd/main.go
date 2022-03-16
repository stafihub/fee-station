// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package main

import (
	"fee-station/pkg/config"
	"fee-station/pkg/db"
	"fee-station/pkg/log"
	"fee-station/pkg/utils"
	"fee-station/task/payer"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/sirupsen/logrus"
	stafihubClient "github.com/stafihub/stafi-hub-relay-sdk/client"
	"os"
	"runtime"
	"runtime/debug"
)

func _main() error {
	cfg, err := config.Load("conf_syncer.toml")
	if err != nil {
		fmt.Printf("loadConfig err: %s", err)
		return err
	}
	log.InitLogFile(cfg.LogFilePath + "/syncer")
	logrus.Infof("config info:%+v ", cfg)

	//init db
	db, err := db.NewDB(&db.Config{
		Host:   cfg.Db.Host,
		Port:   cfg.Db.Port,
		User:   cfg.Db.User,
		Pass:   cfg.Db.Pwd,
		DBName: cfg.Db.Name,
		Mode:   cfg.Mode})
	if err != nil {
		logrus.Errorf("db err: %s", err)
		return err
	}
	logrus.Infof("db connect success")

	//interrupt signal
	ctx := utils.ShutdownListener()
	defer func() {
		sqlDb, err := db.DB.DB()
		if err != nil {
			logrus.Errorf("db.DB() err: %s", err)
			return
		}
		logrus.Infof("shutting down the db ...")
		sqlDb.Close()
	}()

	fmt.Printf("Will open stafihub wallet from <%s>. \nPlease ", cfg.KeystorePath)
	key, err := keyring.New(types.KeyringServiceName(), keyring.BackendFile, cfg.KeystorePath, os.Stdin)
	if err != nil {
		return err
	}
	client, err := stafihubClient.NewClient(key, cfg.PayerAccount, cfg.GasPrice, cfg.StafiHubEndpoint)
	if err != nil {
		return fmt.Errorf("hubClient.NewClient err: %s", err)
	}

	t := task.NewTask(cfg, db, client)
	err = t.Start()
	if err != nil {
		logrus.Errorf("task start err: %s", err)
		return err
	}
	defer func() {
		logrus.Infof("shutting down task ...")
		t.Stop()
	}()

	<-ctx.Done()
	return nil
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	debug.SetGCPercent(40)
	err := _main()
	if err != nil {
		os.Exit(1)
	}
}
