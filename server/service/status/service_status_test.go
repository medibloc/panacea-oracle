package status

import (
	"context"
	"testing"

	"github.com/medibloc/panacea-oracle/mocks"
	"github.com/stretchr/testify/suite"
)

type getStatusTestSuite struct {
	mocks.MockTestSuite
}

func TestGetStatusTestSuite(t *testing.T) {
	suite.Run(t, &getStatusTestSuite{})
}

func (suite *getStatusTestSuite) BeforeTest(_, _ string) {
	suite.Initialize()
}

func (suite *getStatusTestSuite) TestGetStatus() {
	svc := suite.Svc
	oracleAcc := suite.OracleAcc
	conf := suite.Config

	statusService := statusService{
		Service: svc,
	}

	res, err := statusService.GetStatus(context.Background(), nil)
	suite.Require().NoError(err)
	suite.Require().Equal(oracleAcc.GetAddress(), res.OracleAccountAddress)
	suite.Require().Equal(conf.API.Enabled, res.Api.Enabled)
	suite.Require().Equal(conf.API.ListenAddr, res.Api.ListenAddr)
	suite.Require().Equal(conf.GRPC.ListenAddr, res.Grpc.ListenAddr)
	suite.Require().Equal(svc.EnclaveInfo().ProductID, res.EnclaveInfo.ProductId)
	suite.Require().Equal(svc.EnclaveInfo().UniqueIDHex(), res.EnclaveInfo.UniqueId)
}
