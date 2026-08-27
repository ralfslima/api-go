// Pacote principal da aplicação.
// O pacote "main" indica que este arquivo pertence a um programa executável.
package main

// Importações
import (
	// database/sql fornece a interface genérica para trabalhar
	// com bancos de dados relacionais.
	"database/sql"

	// encoding/json permite converter estruturas Go para JSON
	// e JSON para estruturas Go.
	"encoding/json"

	// fmt é utilizado para exibir mensagens no terminal
	// e formatar textos.
	"fmt"

	// log permite registrar mensagens de erro e encerrar
	// a aplicação em situações críticas.
	"log"

	// net/http fornece recursos para criar o servidor HTTP
	// e trabalhar com requisições e respostas.
	"net/http"

	// Biblioteca responsável pela geração de UUIDs.
	"github.com/google/uuid"

	// Driver PostgreSQL utilizado pelo database/sql.
	//
	// O "_" significa que não vamos utilizar diretamente
	// nenhuma função desse pacote no código.
	// O objetivo do import é registrar o driver "postgres".
	_ "github.com/lib/pq"
)

// ============================================================
// MODELO / STRUCT
// ============================================================

// Aluno representa um aluno dentro da aplicação.
//
// Essa struct funciona como o modelo dos dados que serão
// recebidos, enviados e armazenados no banco.
type Aluno struct {

	// Código único do aluno.
	Codigo string `json:"codigo"`

	// Nome do aluno.
	Nome string `json:"nome"`

	// Primeira nota.
	Nota1 float64 `json:"nota1"`

	// Segunda nota.
	Nota2 float64 `json:"nota2"`

	// Média calculada pela aplicação.
	Media float64 `json:"media"`

	// Situação do aluno:
	// Aprovado(a), Em Recuperação ou Reprovado(a).
	Situacao string `json:"situacao"`
}

// ============================================================
// CONEXÃO COM O BANCO
// ============================================================

// Variável global que representa a conexão com o banco.
//
// O *sql.DB não representa necessariamente uma única conexão.
// Ele gerencia um pool de conexões com o banco de dados.
var db *sql.DB

// ============================================================
// INICIALIZAÇÃO DO BANCO DE DADOS
// ============================================================

// initDB inicializa a conexão com o PostgreSQL
// e cria a tabela "alunos", caso ela ainda não exista.
func initDB() {
	// bdURL deve possuir as informações necessárias para realizar a conexão com o banco de dados
	dbURL := "postgresql://banco_api_alunos_06np_user:955WzzG1drWdFXvjkj35LzPWCBxTZZ2m@dpg-da7lnhafngtc73fsj0fg-a/banco_api_alunos_06np"

	// Declara a variável que armazenará possíveis erros.
	var err error

	// Abre a conexão com o PostgreSQL.
	//
	// "postgres" informa qual driver será utilizado.
	db, err = sql.Open("postgres", dbURL)

	// Verifica se houve erro ao abrir/preparar a conexão.
	if err != nil {
		log.Fatalf("Erro ao abrir conexão com o banco: %v", err)
	}

	// Verifica se realmente conseguimos estabelecer
	// comunicação com o banco.
	//
	// sql.Open() sozinho não garante que o banco esteja acessível.
	// O Ping() faz essa verificação.
	err = db.Ping()

	if err != nil {
		log.Fatalf("Erro ao conectar no banco do Render: %v", err)
	}

	// Se chegamos aqui, a conexão foi estabelecida.
	fmt.Println("Conectado ao PostgreSQL com sucesso!")

	// ========================================================
	// CRIAÇÃO DA TABELA
	// ========================================================

	// Comando SQL responsável por criar a tabela.
	//
	// IF NOT EXISTS significa:
	// "crie a tabela somente se ela ainda não existir".
	query := `
	CREATE TABLE IF NOT EXISTS alunos (
		codigo VARCHAR(36) PRIMARY KEY,
		nome VARCHAR(100) NOT NULL,
		nota1 NUMERIC(4,2) NOT NULL,
		nota2 NUMERIC(4,2) NOT NULL,
		media NUMERIC(4,2) NOT NULL,
		situacao VARCHAR(20) NOT NULL
	);`

	// Executa o comando SQL no PostgreSQL.
	_, err = db.Exec(query)

	// Verifica se houve algum problema na criação da tabela.
	if err != nil {
		log.Fatalf("Erro ao criar tabela alunos: %v", err)
	}
}

// ============================================================
// CÁLCULO DA MÉDIA E SITUAÇÃO
// ============================================================

// mediaSituacao calcula a média do aluno
// e determina sua situação.
//
// O parâmetro é um ponteiro (*Aluno), pois queremos
// alterar os valores Media e Situacao do próprio aluno.
func mediaSituacao(aluno *Aluno) {

	// Calcula a média das duas notas.
	aluno.Media = (aluno.Nota1 + aluno.Nota2) / 2

	// Verifica a situação do aluno.
	if aluno.Media >= 7 {

		// Média maior ou igual a 7:
		// aluno aprovado.
		aluno.Situacao = "Aprovado(a)"

	} else if aluno.Media >= 5 {

		// Média entre 5 e 6.99:
		// aluno em recuperação.
		aluno.Situacao = "Em Recuperação"

	} else {

		// Média abaixo de 5:
		// aluno reprovado.
		aluno.Situacao = "Reprovado(a)"
	}
}

// ============================================================
// HELLO WORLD
// ============================================================

// helloWorld é uma função responsável pela rota:
//
// GET /
//
// Ela serve como uma rota simples para testar
// se a API está funcionando.
func helloWorld(w http.ResponseWriter, r *http.Request) {

	// Define que a resposta será JSON.
	w.Header().Set("Content-Type", "application/json")

	// Define o status HTTP da resposta.
	w.WriteHeader(http.StatusCreated)

	// Gera um UUID único.
	codigoUnico := uuid.New().String()

	// Cria um mapa que será convertido para JSON.
	mensagem := map[string]string{
		"codigoUnico": codigoUnico,
		"mensagem":    "Hello World!",
	}

	// Converte o mapa para JSON e envia para o cliente.
	json.NewEncoder(w).Encode(mensagem)
}

// ============================================================
// LISTAR ALUNOS
// ============================================================

// listarAlunos é responsável pela rota:
//
// GET /alunos
//
// Busca todos os alunos cadastrados no PostgreSQL.
func listarAlunos(w http.ResponseWriter, r *http.Request) {

	// Define que a resposta será JSON.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Executa uma consulta SQL para buscar os alunos.
	rows, err := db.Query(`
		SELECT codigo, nome, nota1, nota2, media, situacao
		FROM alunos
	`)

	// Verifica se houve erro na consulta.
	if err != nil {
		http.Error(
			w,
			"Erro ao buscar alunos no banco",
			http.StatusInternalServerError,
		)
		return
	}

	// Garante que as linhas da consulta serão fechadas
	// quando a função terminar.
	defer rows.Close()

	// Criamos uma lista vazia de alunos.
	//
	// Isso garante que, caso não existam alunos,
	// o JSON retornado seja [] em vez de null.
	listaAlunos := []Aluno{}

	// Percorre cada registro retornado pelo banco.
	for rows.Next() {

		// Cria uma variável que representa um aluno.
		var a Aluno

		// Transfere os valores das colunas do banco
		// para os campos da struct Aluno.
		err := rows.Scan(
			&a.Codigo,
			&a.Nome,
			&a.Nota1,
			&a.Nota2,
			&a.Media,
			&a.Situacao,
		)

		// Verifica se houve erro durante a leitura.
		if err != nil {
			http.Error(
				w,
				"Erro ao processar dados dos alunos",
				http.StatusInternalServerError,
			)
			return
		}

		// Adiciona o aluno à lista.
		listaAlunos = append(listaAlunos, a)
	}

	// Define o status HTTP 200 (OK).
	w.WriteHeader(http.StatusOK)

	// Converte a lista de alunos para JSON
	// e envia ao cliente.
	json.NewEncoder(w).Encode(listaAlunos)
}

// ============================================================
// CADASTRAR ALUNO
// ============================================================

// cadastrarAluno é responsável pela rota:
//
// POST /alunos
//
// Recebe os dados de um aluno em JSON
// e salva o aluno no PostgreSQL.
func cadastrarAluno(w http.ResponseWriter, r *http.Request) {

	// Define o formato da resposta.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Cria uma variável para armazenar
	// os dados recebidos no JSON.
	var aluno Aluno

	// Decodifica o JSON enviado pelo cliente
	// diretamente para a struct Aluno.
	erro := json.NewDecoder(r.Body).Decode(&aluno)

	// Verifica se o JSON é inválido.
	if erro != nil {
		http.Error(
			w,
			"Falha ao decodificar o JSON",
			http.StatusBadRequest,
		)
		return
	}

	// Gera automaticamente um código único para o aluno.
	aluno.Codigo = uuid.New().String()

	// Calcula a média e a situação.
	mediaSituacao(&aluno)

	// ========================================================
	// INSERT NO BANCO
	// ========================================================

	// Comando SQL para inserir o aluno.
	//
	// $1, $2, $3 etc. são parâmetros que serão
	// substituídos pelos valores posteriormente.
	query := `
		INSERT INTO alunos
		(codigo, nome, nota1, nota2, media, situacao)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	// Executa o INSERT.
	_, err := db.Exec(
		query,
		aluno.Codigo,
		aluno.Nome,
		aluno.Nota1,
		aluno.Nota2,
		aluno.Media,
		aluno.Situacao,
	)

	// Verifica se houve erro ao salvar.
	if err != nil {
		http.Error(
			w,
			"Erro ao salvar aluno no banco",
			http.StatusInternalServerError,
		)
		return
	}

	// Cadastro realizado com sucesso.
	w.WriteHeader(http.StatusCreated)

	// Retorna o aluno cadastrado em JSON.
	json.NewEncoder(w).Encode(aluno)
}

// ============================================================
// ALTERAR ALUNO
// ============================================================

// alterarAluno é responsável pela rota:
//
// PUT /alunos/{codigo}
//
// Atualiza os dados de um aluno existente.
func alterarAluno(w http.ResponseWriter, r *http.Request) {

	// Define o formato da resposta.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Obtém o código do aluno que está na URL.
	//
	// Exemplo:
	// PUT /alunos/123
	//
	// codigo = "123"
	codigo := r.PathValue("codigo")

	// Cria uma variável para receber
	// os novos dados enviados pelo cliente.
	var aluno Aluno

	// Decodifica o JSON do corpo da requisição.
	erro := json.NewDecoder(r.Body).Decode(&aluno)

	// Verifica se o JSON é inválido.
	if erro != nil {
		http.Error(
			w,
			"Falha ao decodificar o JSON",
			http.StatusBadRequest,
		)
		return
	}

	// Mantém o código original recebido na URL.
	aluno.Codigo = codigo

	// Recalcula a média e a situação
	// com as novas notas.
	mediaSituacao(&aluno)

	// ========================================================
	// UPDATE NO BANCO
	// ========================================================

	// Atualiza os dados do aluno cujo código
	// corresponde ao código recebido na URL.
	query := `
		UPDATE alunos
		SET nome = $1,
			nota1 = $2,
			nota2 = $3,
			media = $4,
			situacao = $5
		WHERE codigo = $6
	`

	// Executa o UPDATE.
	resultado, err := db.Exec(
		query,
		aluno.Nome,
		aluno.Nota1,
		aluno.Nota2,
		aluno.Media,
		aluno.Situacao,
		codigo,
	)

	// Verifica se ocorreu erro no banco.
	if err != nil {
		http.Error(
			w,
			"Erro ao atualizar dados no banco",
			http.StatusInternalServerError,
		)
		return
	}

	// Descobre quantas linhas foram alteradas.
	linhasAfetadas, _ := resultado.RowsAffected()

	// Se nenhuma linha foi alterada,
	// significa que o código não foi encontrado.
	if linhasAfetadas == 0 {
		http.Error(
			w,
			"Código não encontrado",
			http.StatusNotFound,
		)
		return
	}

	// Retorna sucesso.
	w.WriteHeader(http.StatusOK)

	// Retorna o aluno atualizado.
	json.NewEncoder(w).Encode(aluno)
}

// ============================================================
// REMOVER ALUNO
// ============================================================

// removerAluno é responsável pela rota:
//
// DELETE /alunos/{codigo}
//
// Remove um aluno do banco de dados.
func removerAluno(w http.ResponseWriter, r *http.Request) {

	// Define o formato da resposta.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Obtém o código do aluno através da URL.
	codigo := r.PathValue("codigo")

	// ========================================================
	// DELETE NO BANCO
	// ========================================================

	// Remove o aluno que possui o código informado.
	query := "DELETE FROM alunos WHERE codigo = $1"

	// Executa o DELETE.
	resultado, err := db.Exec(query, codigo)

	// Verifica se ocorreu algum erro.
	if err != nil {
		http.Error(
			w,
			"Erro ao remover aluno do banco",
			http.StatusInternalServerError,
		)
		return
	}

	// Verifica quantas linhas foram removidas.
	linhasAfetadas, _ := resultado.RowsAffected()

	// Se nenhuma linha foi removida,
	// o código informado não existe.
	if linhasAfetadas == 0 {
		http.Error(
			w,
			"Código não encontrado",
			http.StatusNotFound,
		)
		return
	}

	// Retorna HTTP 204 (No Content),
	// indicando que a operação foi realizada
	// sem necessidade de retornar um conteúdo.
	w.WriteHeader(http.StatusNoContent)
}

// ============================================================
// CORS
// ============================================================

// corsMiddleware permite que aplicações de outros domínios
// façam requisições para nossa API.
//
// Isso é especialmente importante quando, por exemplo,
// um frontend está hospedado em um endereço diferente
// da API.
func corsMiddleware(next http.Handler) http.Handler {

	// Retorna um novo Handler responsável pelo CORS.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Permite requisições de qualquer origem.
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Define os métodos HTTP permitidos.
		w.Header().Set(
			"Access-Control-Allow-Methods",
			"GET, POST, PUT, DELETE, OPTIONS",
		)

		// Define quais cabeçalhos podem ser enviados.
		w.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type",
		)

		// O navegador pode enviar uma requisição OPTIONS
		// antes da requisição real.
		//
		// Isso é chamado de preflight request.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Se não for OPTIONS, continua normalmente
		// para o próximo Handler.
		next.ServeHTTP(w, r)
	})
}

// ============================================================
// FUNÇÃO PRINCIPAL
// ============================================================

// main é o ponto de entrada da aplicação.
func main() {

	// 1. Inicializa a conexão com o banco
	// antes de iniciar o servidor HTTP.
	initDB()

	// Garante que a conexão com o banco
	// será encerrada quando o programa terminar.
	defer db.Close()

	// ========================================================
	// ROTAS
	// ========================================================

	// GET /
	//
	// Retorna o Hello World.
	http.HandleFunc("GET /", helloWorld)

	// GET /alunos
	//
	// Lista todos os alunos.
	http.HandleFunc("GET /alunos", listarAlunos)

	// POST /alunos
	//
	// Cadastra um novo aluno.
	http.HandleFunc("POST /alunos", cadastrarAluno)

	// PUT /alunos/{codigo}
	//
	// Atualiza um aluno existente.
	http.HandleFunc("PUT /alunos/{codigo}", alterarAluno)

	// DELETE /alunos/{codigo}
	//
	// Remove um aluno.
	http.HandleFunc("DELETE /alunos/{codigo}", removerAluno)

	// ========================================================
	// PORTA
	// ========================================================

	// Definir a porta
	porta := "8080"

	// Exibe no terminal a porta utilizada.
	fmt.Printf("Servidor em execução na porta: %s\n", porta)

	// ========================================================
	// INICIAR SERVIDOR
	// ========================================================

	http.ListenAndServe(
		":"+porta,
		corsMiddleware(http.DefaultServeMux),
	)
}

// // Pacote
// package main

// // Importações
// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/http"

// 	"github.com/google/uuid"
// )

// // Modelo (Struct)
// type Aluno struct {
// 	Codigo   string  `json:"codigo"`
// 	Nome     string  `json:"nome"`
// 	Nota1    float64 `json:"nota1"`
// 	Nota2    float64 `json:"nota2"`
// 	Media    float64 `json:"media"`
// 	Situacao string  `json:"situacao"`
// }

// // Slice
// var alunos = []Aluno{}

// // Função para gerar a média e a situação
// func mediaSituacao(aluno *Aluno) {
// 	// Gerar média
// 	aluno.Media = (aluno.Nota1 + aluno.Nota2) / 2

// 	// Gerar situação
// 	if aluno.Media >= 7 {
// 		aluno.Situacao = "Aprovado(a)"
// 	} else if aluno.Media >= 5 {
// 		aluno.Situacao = "Em Recuperação"
// 	} else {
// 		aluno.Situacao = "Reprovado(a)"
// 	}
// }

// // Função para retornar um Hello World!
// func helloWorld(w http.ResponseWriter, r *http.Request) {
// 	//fmt.Fprintln(w, "Hello World!")

// 	// Definir o cabeçalho
// 	w.Header().Set("Content-Type", "application/json")

// 	// Definir o Status Code
// 	w.WriteHeader(http.StatusCreated)

// 	// Gerar código único
// 	codigoUnico := uuid.New().String()

// 	// Criar JSON mensagem
// 	mensagem := map[string]string{
// 		"codigoUnico": codigoUnico,
// 		"mensagem":    "Hello World!",
// 	}

// 	// Converter o map para JSON e retornar
// 	json.NewEncoder(w).Encode(mensagem)
// }

// // Função responsável pela listagem de alunos
// func listarAlunos(w http.ResponseWriter, r *http.Request) {

// 	// Definir o cabeçalho
// 	w.Header().Set("Content-Type", "application/json; charset=utf-8")

// 	// Definir o Status Code
// 	w.WriteHeader(http.StatusOK)

// 	// Retorna uma lista contendo todos os alunos cadastrados
// 	json.NewEncoder(w).Encode(alunos)
// }

// // Função responsável pelo cadastro de um aluno
// func cadastrarAluno(w http.ResponseWriter, r *http.Request) {

// 	// Definir o cabeçalho
// 	w.Header().Set("Content-Type", "application/json; charset=utf-8")

// 	// Objeto do tipo Aluno
// 	var aluno Aluno

// 	// Decodificar JSON recebido
// 	erro := json.NewDecoder(r.Body).Decode(&aluno)
// 	if erro != nil {
// 		http.Error(w, "Falha ao decodificar o JSON", http.StatusBadRequest)
// 		return
// 	}

// 	// Gerar o código do aluno
// 	aluno.Codigo = uuid.New().String()

// 	// Gerar a média e a situação
// 	mediaSituacao(&aluno)

// 	// Cadastrar no Slice
// 	alunos = append(alunos, aluno)

// 	// Definir o Status Code
// 	w.WriteHeader(http.StatusCreated)

// 	// Retorna o aluno cadastrado
// 	json.NewEncoder(w).Encode(aluno)
// }

// // Função responsável pela alteração de dados de um aluno
// func alterarAluno(w http.ResponseWriter, r *http.Request) {

// 	// Definir o cabeçalho
// 	w.Header().Set("Content-Type", "application/json; charset=utf-8")

// 	// Extrair o código do aluno
// 	codigo := r.PathValue("codigo")

// 	// Laço de repetição
// 	for indice := range alunos {

// 		// Condicional para verificar o código de cada aluno
// 		if alunos[indice].Codigo == codigo {

// 			// Objeto do tipo Aluno
// 			var aluno Aluno

// 			// Decodificar JSON recebido
// 			erro := json.NewDecoder(r.Body).Decode(&aluno)
// 			if erro != nil {
// 				http.Error(w, "Falha ao decodificar o JSON", http.StatusBadRequest)
// 				return
// 			}

// 			// Disponibilizar o código do aluno
// 			aluno.Codigo = codigo

// 			// Gerar a média e a situação
// 			mediaSituacao(&aluno)

// 			// Alterar o aluno no Slice
// 			alunos[indice] = aluno

// 			// Definir o Status Code
// 			w.WriteHeader(http.StatusOK)

// 			// Retorna o aluno com os dados alterados
// 			json.NewEncoder(w).Encode(aluno)

// 			// Finalizar a ação de alteração
// 			return

// 		}

// 	}

// 	// Retorno caso o código informado não exista
// 	http.Error(w, "Código não encontrado", http.StatusNotFound)

// }

// // Função responsável pela remoção de um aluno
// func removerAluno(w http.ResponseWriter, r *http.Request) {

// 	// Definir o cabeçalho
// 	w.Header().Set("Content-Type", "application/json; charset=utf-8")

// 	// Extrair o código do aluno
// 	codigo := r.PathValue("codigo")

// 	// Laço de repetição
// 	for indice := range alunos {

// 		// Condicional para verificar o código de cada aluno
// 		if alunos[indice].Codigo == codigo {

// 			// Remover o aluno no Slice
// 			alunos = append(alunos[:indice], alunos[indice+1:]...)

// 			// Definir o Status Code
// 			w.WriteHeader(http.StatusNoContent)

// 			// Finalizar a ação de remoção
// 			return

// 		}

// 	}

// 	// Retorno caso o código informado não exista
// 	http.Error(w, "Código não encontrado", http.StatusNotFound)

// }

// func corsMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 		// Permitir requisições do frontend
// 		w.Header().Set("Access-Control-Allow-Origin", "*")

// 		// Métodos permitidos
// 		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

// 		// Cabeçalhos permitidos
// 		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

// 		// Tratar requisição preflight
// 		if r.Method == http.MethodOptions {
// 			w.WriteHeader(http.StatusNoContent)
// 			return
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }

// // Função principal
// func main() {

// 	// Rotas
// 	http.HandleFunc("GET /", helloWorld)
// 	http.HandleFunc("GET /alunos", listarAlunos)
// 	http.HandleFunc("POST /alunos", cadastrarAluno)
// 	http.HandleFunc("PUT /alunos/{codigo}", alterarAluno)
// 	http.HandleFunc("DELETE /alunos/{codigo}", removerAluno)

// 	// Retornar o funcionamento do servidor
// 	fmt.Println("Servidor em execução no endereço: http://localhost:8080")

// 	// Configurar servidor
// 	http.ListenAndServe(":8080", corsMiddleware(http.DefaultServeMux))

// }
