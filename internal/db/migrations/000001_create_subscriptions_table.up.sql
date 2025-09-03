CREATE TABLE subscriptions (
    id UUID PRIMARY KEY ,
    service_name VARCHAR(255) NOT NULL,
    price INT NOT NULL,
    user_ID UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE
)