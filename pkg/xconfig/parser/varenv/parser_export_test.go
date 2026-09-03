package varenv

// SetEnvironFuncForTest permite que a nossa suite de testes de caixa-preta
// altere temporariamente a função interna privada environFunc do Parser
func (o *Parser) SetEnvironFuncForTest(fn func() []string) {
	o.environFunc = fn
}
