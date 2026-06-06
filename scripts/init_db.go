package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = "root:root@tcp(127.0.0.1:3306)/?charset=utf8mb4&parseTime=True"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("connect error:", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Println("ping error:", err)
		os.Exit(1)
	}

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS `ns-tracking` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
	if err != nil {
		fmt.Println("create db error:", err)
		os.Exit(1)
	}
	fmt.Println("Database ns-tracking ready")

	_, err = db.Exec("USE `ns-tracking`")
	if err != nil {
		fmt.Println("use db error:", err)
		os.Exit(1)
	}

	// raw_events
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS raw_events (
		id              BIGINT AUTO_INCREMENT PRIMARY KEY,
		idempotency_key VARCHAR(255) NOT NULL UNIQUE,
		provider_code   VARCHAR(50)  NOT NULL,
		data_code       VARCHAR(50)  NOT NULL,
		waybill_number  VARCHAR(100),
		tracking_number VARCHAR(100),
		customer_code   VARCHAR(100),
		track_node_code VARCHAR(100),
		process_time    VARCHAR(50),
		payload         TEXT         NOT NULL,
		envelope_meta   TEXT,
		status          VARCHAR(20)  NOT NULL DEFAULT 'pending',
		retry_count     INT          NOT NULL DEFAULT 0,
		max_retries     INT          NOT NULL DEFAULT 5,
		last_error      TEXT,
		processed_at    TIMESTAMP    NULL,
		created_at      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		fmt.Println("create raw_events error:", err)
		os.Exit(1)
	}
	fmt.Println("Table raw_events ready")

	// tracking_details
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS tracking_details (
		id                BIGINT AUTO_INCREMENT PRIMARY KEY,
		tracking_number   VARCHAR(100) NOT NULL,
		service_class     VARCHAR(100),
		status            INT          NOT NULL DEFAULT 0,
		detail            JSON,
		last_detail       JSON,
		synced_at         TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
		error_message     TEXT,
		tracking_log_id   BIGINT,
		auto_delivered_at TIMESTAMP    NULL,
		created_at        TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at        TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY uk_tracking_number (tracking_number)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		fmt.Println("create tracking_details error:", err)
		os.Exit(1)
	}
	fmt.Println("Table tracking_details ready")

	// tracking_logs (simplified — read from existing Ruby DB)
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS tracking_logs (
		id                     BIGINT AUTO_INCREMENT PRIMARY KEY,
		tracking_number        VARCHAR(100),
		source_tracking_number VARCHAR(100),
		channel_alias          VARCHAR(100),
		shipping_agent         VARCHAR(100),
		shipping_channel       VARCHAR(100),
		country_code           VARCHAR(10),
		track_status           INT          NOT NULL DEFAULT 0,
		synced_at              TIMESTAMP    NULL,
		received_at            TIMESTAMP    NULL,
		delivered_at           TIMESTAMP    NULL,
		tracked_at             TIMESTAMP    NULL,
		fulfill_at             TIMESTAMP    NULL,
		created_at             TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at             TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		fmt.Println("create tracking_logs error:", err)
		os.Exit(1)
	}
	fmt.Println("Table tracking_logs ready")

	fmt.Println("All tables created successfully!")
}
