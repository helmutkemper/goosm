package passwordHash

import (
	"errors"
	constants "github.com/helmutkemper/kemper.com.br.plugin.dataaccess.constants"
	"log"
)

func (e *Password) ruleLength(password []byte) (err error) {
	if len(password) < 8 {
		err = errors.New(constants.KErrorPasswordMustBe8LettersOrMore)
		log.Printf("passwordHash.ruleLength().error: %v", err.Error())
	}

	return
}
