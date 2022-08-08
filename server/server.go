// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package server

import (
	"fee-station/api"
	dao_station "fee-station/dao/station"
	"fee-station/pkg/config"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	hubClient "github.com/stafihub/cosmos-relay-sdk/client"
	"gorm.io/gorm"
)

type Server struct {
	listenAddr string
	httpServer *http.Server
	taskTicker int64
	cfg        *config.Config
	db         *db.WrapDb
}

func NewServer(cfg *config.Config, dao *db.WrapDb) (*Server, error) {
	s := &Server{
		listenAddr: cfg.ListenAddr,
		taskTicker: cfg.TaskTicker,
		cfg:        cfg,
		db:         dao,
	}

	cache := map[string]string{}

	handler := s.InitHandler(cache)

	s.httpServer = &http.Server{
		Addr:         s.listenAddr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return s, nil
}

func (svr *Server) InitHandler(cache map[string]string) http.Handler {
	return api.InitRouters(svr.db, cache)
}

func (svr *Server) ApiServer() {
	logrus.Infof("Gin server start on %s", svr.listenAddr)
	err := svr.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logrus.Errorf("Gin server start err: %s", err.Error())
		utils.ShutdownRequestChannel <- struct{}{} //shutdown server
		return
	}
	logrus.Infof("Gin server done on %s", svr.listenAddr)
}

func (svr *Server) InitOrUpdatePoolAddress() error {
	for _, tokenInfo := range svr.cfg.TokenInfo {
		client, err := hubClient.NewClient(nil, "", "", tokenInfo.AccountPrefix, []string{tokenInfo.Endpoint})
		if err != nil {
			return err
		}
		res, err := client.QueryBondedDenom()
		if err != nil {
			return err
		}

		metaData, err := dao_station.GetMetaData(svr.db, res.Params.BondDenom)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		metaData.Symbol = res.Params.BondDenom
		metaData.AccountPrefix = tokenInfo.AccountPrefix
		metaData.CoinmarketSymbol = tokenInfo.CoinMarketSymbol
		metaData.CoinGeckoSymbol = tokenInfo.CoinGeckoSymbol
		metaData.PoolAddress = tokenInfo.PoolAddress
		metaData.Endpoint = tokenInfo.Endpoint
		metaData.Decimals = tokenInfo.Decimals
		err = dao_station.UpOrInMetaData(svr.db, metaData)
		if err != nil {
			return err
		}
	}

	limitInfo, err := dao_station.GetLimitInfo(svr.db)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	limitInfo.SwapMaxLimit = svr.cfg.SwapMaxLimit
	limitInfo.SwapMinLimit = svr.cfg.SwapMinLimit
	limitInfo.SwapRate = svr.cfg.SwapRate

	return dao_station.UpOrInLimitInfo(svr.db, limitInfo)

}

func (svr *Server) Start() error {
	err := svr.InitOrUpdatePoolAddress()
	if err != nil {
		return err
	}
	utils.SafeGoWithRestart(svr.ApiServer)
	return nil
}

func (svr *Server) Stop() {
	if svr.httpServer != nil {
		err := svr.httpServer.Close()
		if err != nil {
			logrus.Errorf("Problem shutdown Gin server :%s", err.Error())
		}
	}
}
