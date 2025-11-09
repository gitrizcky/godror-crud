package repo

import (
    "database/sql"

    cfg "go-demo-crud/internal/config"
    "go-demo-crud/internal/model"
)

var ErrNotFound = sql.ErrNoRows

type ProductRepo struct {
    DB *sql.DB
    Q  *cfg.Manager
}

func NewProductRepo(db *sql.DB, q *cfg.Manager) *ProductRepo {
    return &ProductRepo{DB: db, Q: q}
}

func (r *ProductRepo) ListProducts() ([]model.Product, error) {
    rows, err := r.DB.Query(r.Q.Get().ListProducts)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    items := make([]model.Product, 0)
    for rows.Next() {
        var (
            id int64
            name string
            attrV sql.NullString
            attrC sql.NullString
            attrB []byte
        )
        if err := rows.Scan(&id, &name, &attrV, &attrC, &attrB); err != nil {
            return nil, err
        }
        p := model.Product{ProductID: id, Name: name, AttributesBlob: attrB}
        if attrV.Valid {
            v := attrV.String
            p.AttributesVarchar = &v
        }
        if attrC.Valid {
            c := attrC.String
            p.AttributesClob = &c
        }
        items = append(items, p)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    return items, nil
}

func (r *ProductRepo) GetProduct(id int64) (*model.Product, error) {
    var (
        name string
        attrV sql.NullString
        attrC sql.NullString
        attrB []byte
    )
    err := r.DB.QueryRow(r.Q.Get().GetProduct, id).
        Scan(&name, &attrV, &attrC, &attrB)
    if err != nil {
        return nil, err
    }
    p := &model.Product{ProductID: id, Name: name, AttributesBlob: attrB}
    if attrV.Valid {
        v := attrV.String
        p.AttributesVarchar = &v
    }
    if attrC.Valid {
        c := attrC.String
        p.AttributesClob = &c
    }
    return p, nil
}

func (r *ProductRepo) nextProductID(tx *sql.Tx) (int64, error) {
    var id int64
    err := tx.QueryRow(r.Q.Get().NextProductID).Scan(&id)
    return id, err
}

func (r *ProductRepo) CreateProduct(in model.Product) (*model.Product, error) {
    tx, err := r.DB.Begin()
    if err != nil {
        return nil, err
    }
    defer func() { _ = tx.Rollback() }()

    id, err := r.nextProductID(tx)
    if err != nil {
        return nil, err
    }

    _, err = tx.Exec(
        r.Q.Get().InsertProduct,
        id,
        in.Name,
        in.AttributesVarchar,
        in.AttributesClob,
        in.AttributesBlob,
    )
    if err != nil {
        return nil, err
    }
    if err := tx.Commit(); err != nil {
        return nil, err
    }
    in.ProductID = id
    return &in, nil
}

func (r *ProductRepo) UpdateProduct(id int64, in model.Product) (bool, error) {
    res, err := r.DB.Exec(
        r.Q.Get().UpdateProduct,
        in.Name,
        in.AttributesVarchar,
        in.AttributesClob,
        in.AttributesBlob,
        id,
    )
    if err != nil {
        return false, err
    }
    n, _ := res.RowsAffected()
    return n > 0, nil
}

func (r *ProductRepo) DeleteProduct(id int64) (bool, error) {
    res, err := r.DB.Exec(r.Q.Get().DeleteProduct, id)
    if err != nil {
        return false, err
    }
    n, _ := res.RowsAffected()
    return n > 0, nil
}
