-- +migrate Up
CREATE TABLE IF NOT EXISTS holiday_category_links (
    holiday_id INT REFERENCES holidays(id) ON DELETE CASCADE,
    category_id INT REFERENCES holiday_categories(id) ON DELETE CASCADE,
    PRIMARY KEY (holiday_id, category_id)
);