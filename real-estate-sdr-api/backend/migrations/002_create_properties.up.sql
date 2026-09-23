CREATE TABLE properties (
                            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                            title VARCHAR(200) NOT NULL,
                            description TEXT NOT NULL,

                            property_type VARCHAR(30) NOT NULL,
                            transaction_type VARCHAR(20) NOT NULL,

                            price NUMERIC(14,2) NOT NULL,

                            bedrooms INTEGER NOT NULL DEFAULT 0,
                            bathrooms INTEGER NOT NULL DEFAULT 0,
                            parking_spaces INTEGER NOT NULL DEFAULT 0,

                            area NUMERIC(10,2) NOT NULL,

                            city VARCHAR(100) NOT NULL,
                            neighborhood VARCHAR(100) NOT NULL,
                            address VARCHAR(255) NOT NULL,

                            available BOOLEAN NOT NULL DEFAULT TRUE,

                            created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
                            updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

                            CONSTRAINT ck_properties_type
                                CHECK (
                                    property_type IN (
                                                      'APARTMENT',
                                                      'HOUSE',
                                                      'COMMERCIAL',
                                                      'LAND'
                                        )
                                    ),

                            CONSTRAINT ck_properties_transaction
                                CHECK (
                                    transaction_type IN (
                                                         'SALE',
                                                         'RENT'
                                        )
                                    ),

                            CONSTRAINT ck_properties_price
                                CHECK (price >= 0),

                            CONSTRAINT ck_properties_bedrooms
                                CHECK (bedrooms >= 0),

                            CONSTRAINT ck_properties_bathrooms
                                CHECK (bathrooms >= 0),

                            CONSTRAINT ck_properties_parking
                                CHECK (parking_spaces >= 0),

                            CONSTRAINT ck_properties_area
                                CHECK (area > 0)
);

CREATE INDEX idx_properties_city
    ON properties(city);

CREATE INDEX idx_properties_neighborhood
    ON properties(neighborhood);

CREATE INDEX idx_properties_type
    ON properties(property_type);

CREATE INDEX idx_properties_transaction_type
    ON properties(transaction_type);

CREATE INDEX idx_properties_price
    ON properties(price);

CREATE INDEX idx_properties_bedrooms
    ON properties(bedrooms);

CREATE INDEX idx_properties_available
    ON properties(available);