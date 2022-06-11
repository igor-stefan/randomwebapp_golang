package dbrepo

import (
	"errors"
	"time"

	"github.com/igor-stefan/myfirstwebapp_golang/internal/models"
)

func (m *testPostgresDBRepo) AllUsers() bool {
	return true
}

// InsertReservation acrecenta uma linha na tabela 'Reserva' do db
func (m *testPostgresDBRepo) InsertReserva(res models.Reserva) (int, error) {
	if res.LivroID == -1 {
		return res.LivroID, errors.New("numero do livro invalido, nao foi possivel inserir a reserva no db")
	}
	return res.LivroID, nil
}

// InsertLivroRestricao insere uma nova restricao para determinado livro no db
func (m *testPostgresDBRepo) InsertLivroRestricao(r models.LivroRestricao) error {
	if r.LivroID == -2 {
		return errors.New("numero do livro invalido, nao foi possivel inserir a restricao do livro no db")
	}
	return nil
}

// SearchAvailabilityByDatesByRoomID retorna true se existe disponibilidade
func (m *testPostgresDBRepo) SearchAvailabilityByDatesByLivroID(inicio, fim time.Time, livroID int) (bool, error) {
	if livroID == -1 {
		return false, nil
	} else if livroID == -2 {
		return false, errors.New("erro ao conectar-se ao db")
	}
	return true, nil
}

// SearchAvailabilityForAllLivros retorna um slice de livros que estao disponiveis para as datas especificados
func (m *testPostgresDBRepo) SearchAvailabilityForAllLivros(inicio, final time.Time) ([]models.Livro, error) {
	var livros []models.Livro
	return livros, nil
}

// GetLivroByID busca no database o livro que possui o id especificado
func (m *testPostgresDBRepo) GetLivroByID(ID int) (models.Livro, error) {
	var livro models.Livro
	if ID > 7 {
		return livro, errors.New("nao foi encontrado livro com ID especificado")
	}
	return livro, nil
}