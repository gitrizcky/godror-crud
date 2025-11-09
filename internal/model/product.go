package model

type Product struct {
    ProductID         int64   `json:"product_id"`
    Name              string  `json:"name"`
    AttributesVarchar *string `json:"attributes_varchar,omitempty"`
    AttributesClob    *string `json:"attributes_clob,omitempty"`
    AttributesBlob    []byte  `json:"attributes_blob,omitempty"`
}

