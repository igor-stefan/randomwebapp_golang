package dbrepo

import (
	"context"
	"errors"
	"time"

	"github.com/igor-stefan/myfirstwebapp_golang/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

// InsertReservation acrecenta uma linha na tabela 'Reserva' do db
func (m *postgresDBRepo) InsertReserva(res models.Reserva) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //cria um contexto para evitar que a solicitacao demore mais que 3 seg
	defer cancel()

	var returnedID int
	stmt := `insert into 
			reservas (nome, sobrenome, email, phone, data_inicio, data_final, livro_id, created_at, updated_at)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`
	err := m.DB.QueryRowContext(ctx, stmt,
		res.Nome,
		res.Sobrenome,
		res.Email,
		res.Phone,
		res.DataInicio,
		res.DataFinal,
		res.LivroID,
		time.Now(),
		time.Now(),
	).Scan(&returnedID)
	if err != nil {
		return -1, err
	}
	return returnedID, nil
}

// InsertLivroRestricao insere uma nova restricao para determinado livro no db
func (m *postgresDBRepo) InsertLivroRestricao(r models.LivroRestricao) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //cria um contexto para evitar que a solicitacao demore mais que 3 seg
	defer cancel()

	stmt := `insert into 
	livros_restricoes (data_inicio, data_final, id_livro, id_reserva, created_at, updated_at, id_restricao)
	values ($1, $2, $3, $4, $5, $6, $7)`

	_, err := m.DB.ExecContext(ctx, stmt,
		r.DataInicio,
		r.DataFinal,
		r.LivroID,
		r.ReservaID,
		time.Now(),
		time.Now(),
		r.RestricaoID,
	)
	if err != nil {
		return err
	}
	return nil
}

// SearchAvailabilityByDatesByLivroID retorna true se existe disponibilidade
func (m *postgresDBRepo) SearchAvailabilityByDatesByLivroID(inicio, fim time.Time, livroID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //cria um contexto para evitar que a solicitacao demore mais que 3 seg
	defer cancel()

	query := `select count(id)
		from livros_restricoes
		where id_livro = $1 and $2 <= data_final and $3 >= data_inicio;`

	var numRows int
	row := m.DB.QueryRowContext(ctx, query, livroID, inicio, fim)
	err := row.Scan(&numRows)
	if err != nil {
		return false, err
	}
	if numRows == 0 {
		return true, nil
	} else {
		return false, nil
	}
}

// SearchAvailabilityForAllLivros retorna um slice de livros que estao disponiveis para as datas especificados
func (m *postgresDBRepo) SearchAvailabilityForAllLivros(inicio, final time.Time) ([]models.Livro, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //cria um contexto para evitar que a solicitacao demore mais que 3 seg
	defer cancel()

	var livros []models.Livro

	query := `select l.id_livro, l.nome_livro
	from livros l
	where l.id_livro not in
	(select id_livro from livros_restricoes lr where $1 <= lr.data_final and $2 >= lr.data_inicio);`

	rows, err := m.DB.QueryContext(ctx, query, inicio, final)
	if err != nil {
		return livros, err
	}
	defer rows.Close()

	for rows.Next() {
		var livro models.Livro
		err := rows.Scan(
			&livro.ID,
			&livro.NomeLivro,
		)
		if err != nil {
			return livros, err
		}
		livros = append(livros, livro)
	}

	if err = rows.Err(); err != nil {
		return livros, err
	}
	return livros, nil
}

// GetLivroByID busca no database o livro que possui o id especificado
func (m *postgresDBRepo) GetLivroByID(ID int) (models.Livro, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //cria um contexto para evitar que a solicitacao demore mais que 3 seg
	defer cancel()

	var livro models.Livro

	query := `select id_livro, nome_livro, created_at, updated_at from livros where id_livro = $1`
	row := m.DB.QueryRowContext(ctx, query, ID)
	err := row.Scan(
		&livro.ID,
		&livro.NomeLivro,
		&livro.CreatedAt,
		&livro.UpdatedAt,
	)
	if err != nil {
		return livro, err
	}
	return livro, nil
}

// GetUserById retorna dados de um usuário especificado pelo seu ID
func (m *postgresDBRepo) GetUserById(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, nome, sobrenome, email, senha, acces_level, created_at, updated_at
	from users where id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	var u models.User
	err := row.Scan(
		&u.Id,
		&u.Nome,
		&u.SobreNome,
		&u.Senha,
		&u.AccessLevel,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return u, err
	}
	return u, nil
}

// UpdateUser altera dados de um usuario no db
func (m *postgresDBRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update users set nome = $1, sobrenome = $2, email = $3, acces_level = $4, updated_at = $5`

	_, err := m.DB.ExecContext(ctx, query, u.Nome, u.SobreNome, u.Email, u.AccessLevel, time.Now())
	if err != nil {
		return err
	}
	return nil
}

// Autenticar retorna id e senha em caso de sucesso na autenticacao do usuário
func (m *postgresDBRepo) Autenticar(email, senhaFornecida string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedSenha string

	row := m.DB.QueryRowContext(ctx, "select id, senha from users where email = $1", email)

	err := row.Scan(&id, &hashedSenha)
	if err != nil {
		return id, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedSenha), []byte(senhaFornecida))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return -1, "", errors.New("senha incorreta")
	} else if err != nil {
		return -1, "", err
	}
	return id, hashedSenha, nil
}

// AllReservas retorna todas as reservas presentes no db
func (m *postgresDBRepo) AllReservas() ([]models.Reserva, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservas []models.Reserva
	query := `select r.id, r.nome, r.sobrenome, r.email, r.phone, r.data_inicio, r.data_final, r.livro_id,
	r.created_at, r.updated_at, lv.id_livro, lv.nome_livro 
	from reservas r
	left join livros lv on (r.livro_id  = lv.id_livro)
	order by r.id`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservas, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Reserva
		err := rows.Scan(
			&item.ID,
			&item.Nome,
			&item.Sobrenome,
			&item.Email,
			&item.Phone,
			&item.DataInicio,
			&item.DataFinal,
			&item.LivroID,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.Livro.ID,
			&item.Livro.NomeLivro)
		if err != nil {
			return reservas, err
		}
		reservas = append(reservas, item)
	}
	if err = rows.Err(); err != nil {
		return reservas, err
	}
	return reservas, nil
}

// NewReservas lista todas as reservas não processadas
func (m *postgresDBRepo) NewReservas() ([]models.Reserva, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservas []models.Reserva
	query := `select r.id, r.nome, r.sobrenome, r.email, r.phone, r.data_inicio, r.data_final, r.livro_id,
	r.created_at, r.updated_at, r.processada, lv.id_livro, lv.nome_livro 
	from reservas r
	left join livros lv on (r.livro_id  = lv.id_livro)
	where processada = 0
	order by r.id`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservas, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Reserva
		err := rows.Scan(
			&item.ID,
			&item.Nome,
			&item.Sobrenome,
			&item.Email,
			&item.Phone,
			&item.DataInicio,
			&item.DataFinal,
			&item.LivroID,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.Processada,
			&item.Livro.ID,
			&item.Livro.NomeLivro)
		if err != nil {
			return reservas, err
		}
		reservas = append(reservas, item)
	}
	if err = rows.Err(); err != nil {
		return reservas, err
	}
	return reservas, nil
}

// GetReservaById retorna a reserva com o id específico
func (m *postgresDBRepo) GetReservaById(id int) (models.Reserva, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var res models.Reserva
	query := `select r.id, r.nome, r.sobrenome, r.email, r.phone, r.data_inicio, r.data_final, r.livro_id, r.created_at, r.updated_at, r.processada,
	lv.id_livro, lv.nome_livro 
	from reservas r 
	left join livros lv on (r.livro_id = lv.id_livro)
	where id = $1`
	row := m.DB.QueryRowContext(ctx, query, id) // query ao db
	err := row.Scan(
		&res.ID,
		&res.Nome,
		&res.Sobrenome,
		&res.Email,
		&res.Phone,
		&res.DataInicio,
		&res.DataFinal,
		&res.Livro.ID,
		&res.CreatedAt,
		&res.UpdatedAt,
		&res.Processada,
		&res.Livro.ID,
		&res.Livro.NomeLivro)
	if err != nil {
		return res, err
	}
	return res, nil
}

// UpdateReservas altera dados de uma reserva no db
func (m *postgresDBRepo) UpdateReserva(r models.Reserva) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update reservas set nome = $1, sobrenome = $2, phone = $3, email = $4, updated_at = $5
	where id = $6`

	_, err := m.DB.ExecContext(ctx, query, r.Nome, r.Sobrenome, r.Phone, r.Email, time.Now(), r.ID)
	if err != nil {
		return err
	}
	return nil
}

// DeleteReserva deleta uma reserva específica do db, encontrado pelo parametro id fornecido
func (m *postgresDBRepo) DeleteReserva(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `delete from reservas where id = $1`
	_, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}

// UpdateProcessadaForReserva altera o estado de processada para uma reserva
func (m *postgresDBRepo) UpdateProcessadaForReserva(id, processada int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update reservas set processada = $1 where id = $2`
	_, err := m.DB.ExecContext(ctx, query, processada, id)
	if err != nil {
		return err
	}
	return nil
}

// AllLivros retorna todos os livros armazenados no db
func (m *postgresDBRepo) AllLivros() ([]models.Livro, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var livRet []models.Livro
	query := "select l.id_livro, l.nome_livro, l.num_emprestimos, l.created_at, l.updated_at from livros l"
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return livRet, err
	}
	defer rows.Close()

	for rows.Next() {
		var livAt models.Livro
		err = rows.Scan(
			&livAt.ID,
			&livAt.NomeLivro,
			&livAt.Emprestimos,
			&livAt.CreatedAt,
			&livAt.UpdatedAt,
		)
		if err != nil {
			return livRet, err
		}
		livRet = append(livRet, livAt)
	}

	if err = rows.Err(); err != nil {
		return livRet, err
	}
	return livRet, nil
}

// GetRestricoesForLivroByDate retorna possiveis restricoes para um determinado livro entre duas datas
func (m *postgresDBRepo) GetRestricoesForLivroByDate(id_livro int, inicio, final time.Time) ([]models.LivroRestricao, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var restricoes []models.LivroRestricao
	query := `select l.id, l.id_livro, coalesce(l.id_reserva, 0), l.id_restricao, l.data_inicio, l.data_final, l.created_at, l.updated_at 
	from livros_restricoes l
	where $1 <= l.data_final and $2 >= l.data_inicio and l.id_livro = $3`
	rows, err := m.DB.QueryContext(ctx, query, inicio, final, id_livro)
	if err != nil {
		return restricoes, err
	}
	defer rows.Close()

	for rows.Next() {
		var rest models.LivroRestricao
		err = rows.Scan(
			&rest.ID,
			&rest.LivroID,
			&rest.ReservaID,
			&rest.RestricaoID,
			&rest.DataInicio,
			&rest.DataFinal,
			&rest.CreatedAt,
			&rest.UpdatedAt,
		)
		if err != nil {
			return restricoes, err
		}
		restricoes = append(restricoes, rest)
	}

	if err = rows.Err(); err != nil {
		return restricoes, err
	}
	return restricoes, nil
}

func (m *postgresDBRepo) InsertBlockForLivro(id int, dataInicio time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ordem := `insert into 
	livros_restricoes (data_inicio, data_final, id_livro, id_reserva, id_restricao, created_at, updated_at)
	values ($1, $2, $3, $4, $5, $6, $7)`

	_, err := m.DB.ExecContext(ctx, ordem, dataInicio, dataInicio, id, 0, 2, time.Now(), time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (m *postgresDBRepo) DeleteBlockForLivro(id int, dataInicio time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ordem := `delete from livros_restricoes where id_livro = $1 and data_inicio = $2 and data_final = $3`
	_, err := m.DB.ExecContext(ctx, ordem, id, dataInicio, dataInicio)
	if err != nil {
		return err
	}
	return nil
}
