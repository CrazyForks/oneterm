package service

import (
	"fmt"

	"golang.org/x/crypto/ssh"

	"github.com/veops/oneterm/internal/model"
	dbpkg "github.com/veops/oneterm/pkg/db"
	"github.com/veops/oneterm/pkg/utils"
)

func GetAAG(assetId int, accountId int) (asset *model.Asset, account *model.Account, gateway *model.Gateway, err error) {
	asset, account, gateway = &model.Asset{}, &model.Account{}, &model.Gateway{}
	if err = dbpkg.DB.Model(asset).Where("id = ?", assetId).First(asset).Error; err != nil {
		return
	}
	if err = dbpkg.DB.Model(account).Where("id = ?", accountId).First(account).Error; err != nil {
		return
	}
	account.Password = utils.DecryptAES(account.Password)
	account.Pk = utils.DecryptAES(account.Pk)
	account.Phrase = utils.DecryptAES(account.Phrase)
	if asset.GatewayId != 0 {
		if err = dbpkg.DB.Model(gateway).Where("id = ?", asset.GatewayId).First(gateway).Error; err != nil {
			return
		}
		gateway.Password = utils.DecryptAES(gateway.Password)
		gateway.Pk = utils.DecryptAES(gateway.Pk)
		gateway.Phrase = utils.DecryptAES(gateway.Phrase)
	}

	return
}

func GetAuth(account *model.Account) (ssh.AuthMethod, error) {
	switch account.AccountType {
	case model.AUTHMETHOD_PASSWORD:
		return ssh.Password(account.Password), nil
	case model.AUTHMETHOD_PUBLICKEY:
		if account.Phrase == "" {
			pk, err := ssh.ParsePrivateKey([]byte(account.Pk))
			if err != nil {
				return nil, err
			}
			return ssh.PublicKeys(pk), nil
		} else {
			pk, err := ssh.ParsePrivateKeyWithPassphrase([]byte(account.Pk), []byte(account.Phrase))
			if err != nil {
				return nil, err
			}
			return ssh.PublicKeys(pk), nil
		}
	default:
		return nil, fmt.Errorf("invalid authmethod %d", account.AccountType)
	}
}
