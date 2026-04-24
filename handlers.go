package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type App struct {
	DB *sql.DB
}

func (a *App) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *App) seriesCollectionHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.listSeries(w, r)
	case http.MethodPost:
		a.createSeries(w, r)
	default:
		a.writeJSONError(w, http.StatusNotFound, "endpoint not found")
	}
}

func (a *App) seriesItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		a.writeJSONError(w, http.StatusNotFound, "series not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		a.getSeries(w, r, id)
	case http.MethodPut:
		a.updateSeries(w, r, id)
	case http.MethodDelete:
		a.deleteSeries(w, r, id)
	default:
		a.writeJSONError(w, http.StatusNotFound, "endpoint not found")
	}
}

func (a *App) listSeries(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	page := parsePositiveIntOrDefault(query.Get("page"), 1)
	limit := parsePositiveIntOrDefault(query.Get("limit"), 5)
	if limit > 100 {
		limit = 100
	}

	search := strings.TrimSpace(query.Get("q"))
	sortField := query.Get("sort")
	if sortField == "" {
		sortField = "id"
	}

	allowedSorts := map[string]string{
		"id":       "id",
		"title":    "title",
		"episodes": "episodes",
	}
	sortColumn, ok := allowedSorts[sortField]
	if !ok {
		a.writeJSONError(w, http.StatusBadRequest, "invalid sort field")
		return
	}

	order := strings.ToUpper(query.Get("order"))
	if order == "" {
		order = "ASC"
	}
	if order != "ASC" && order != "DESC" {
		a.writeJSONError(w, http.StatusBadRequest, "invalid order value")
		return
	}

	where := ""
	args := []any{}
	if search != "" {
		where = "WHERE title LIKE ? OR description LIKE ?"
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM series %s", where)
	if err := a.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		a.writeJSONError(w, http.StatusBadRequest, "failed to count series")
		return
	}

	offset := (page - 1) * limit
	listQuery := fmt.Sprintf(`
		SELECT id, title, description, episodes, image
		FROM series
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?`,
		where,
		sortColumn,
		order,
	)

	listArgs := append(append([]any{}, args...), limit, offset)
	rows, err := a.DB.Query(listQuery, listArgs...)
	if err != nil {
		a.writeJSONError(w, http.StatusBadRequest, "failed to fetch series")
		return
	}
	defer rows.Close()

	seriesList := make([]Series, 0)
	for rows.Next() {
		var item Series
		if err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.Episodes, &item.Image); err != nil {
			a.writeJSONError(w, http.StatusBadRequest, "failed to read series")
			return
		}
		seriesList = append(seriesList, item)
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	response := SeriesListResponse{
		Data:       seriesList,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
	a.writeJSON(w, http.StatusOK, response)
}

func (a *App) getSeries(w http.ResponseWriter, _ *http.Request, id int) {
	var item Series
	err := a.DB.QueryRow(`
		SELECT id, title, description, episodes, image
		FROM series
		WHERE id = ?`,
		id,
	).Scan(&item.ID, &item.Title, &item.Description, &item.Episodes, &item.Image)
	if err == sql.ErrNoRows {
		a.writeJSONError(w, http.StatusNotFound, "series not found")
		return
	}
	if err != nil {
		a.writeJSONError(w, http.StatusBadRequest, "failed to fetch series")
		return
	}

	a.writeJSON(w, http.StatusOK, item)
}

func (a *App) createSeries(w http.ResponseWriter, r *http.Request) {
	var input Series
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		a.writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := validateSeries(input); err != nil {
		a.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := a.DB.Exec(`
		INSERT INTO series (title, description, episodes, image)
		VALUES (?, ?, ?, ?)`,
		input.Title,
		input.Description,
		input.Episodes,
		input.Image,
	)
	if err != nil {
		a.writeJSONError(w, http.StatusBadRequest, "failed to create series")
		return
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		a.writeJSONError(w, http.StatusBadRequest, "failed to read created series")
		return
	}

	input.ID = int(lastID)
	a.writeJSON(w, http.StatusCreated, input)
}

func (a *App) updateSeries(w http.ResponseWriter, r *http.Request, id int) {
	var existingID int
	err := a.DB.QueryRow("SELECT id FROM series WHERE id = ?", id).Scan(&existingID)
	if err == sql.ErrNoRows {
		a.writeJSONError(w, http.StatusNotFound, "series not found")
		return
	}
	if err != nil {
		a.writeJSONError(w, http.StatusBadRequest, "failed to fetch series")
		return
	}

	var input Series
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		a.writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := validateSeries(input); err != nil {
		a.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := a.DB.Exec(`
		UPDATE series
		SET title = ?, description = ?, episodes = ?, image = ?
		WHERE id = ?`,
		input.Title,
		input.Description,
		input.Episodes,
		input.Image,
		id,
	); err != nil {
		a.writeJSONError(w, http.StatusBadRequest, "failed to update series")
		return
	}

	input.ID = id
	a.writeJSON(w, http.StatusOK, input)
}

func (a *App) deleteSeries(w http.ResponseWriter, _ *http.Request, id int) {
	result, err := a.DB.Exec("DELETE FROM series WHERE id = ?", id)
	if err != nil {
		a.writeJSONError(w, http.StatusBadRequest, "failed to delete series")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		a.writeJSONError(w, http.StatusBadRequest, "failed to delete series")
		return
	}
	if rowsAffected == 0 {
		a.writeJSONError(w, http.StatusNotFound, "series not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) swaggerJSONHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write([]byte(swaggerDocument))
}

func (a *App) swaggerUIHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Series API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function () {
      window.SwaggerUIBundle({
        url: '/swagger.json',
        dom_id: '#swagger-ui'
      });
    };
  </script>
</body>
</html>`))
}

func (a *App) healthHandler(w http.ResponseWriter, _ *http.Request) {
	a.writeJSON(w, http.StatusOK, map[string]string{"message": "Series API is running"})
}

func (a *App) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (a *App) writeJSONError(w http.ResponseWriter, status int, message string) {
	a.writeJSON(w, status, ErrorResponse{Error: message})
}

func parseID(path string) (int, error) {
	trimmed := strings.Trim(path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 || parts[0] != "series" {
		return 0, fmt.Errorf("invalid path")
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return id, nil
}

func parsePositiveIntOrDefault(value string, fallback int) int {
	if value == "" {
		return fallback
	}

	number, err := strconv.Atoi(value)
	if err != nil || number <= 0 {
		return fallback
	}

	return number
}

func validateSeries(input Series) error {
	if strings.TrimSpace(input.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if strings.TrimSpace(input.Description) == "" {
		return fmt.Errorf("description is required")
	}
	if input.Episodes <= 0 {
		return fmt.Errorf("episodes must be a positive integer")
	}
	if strings.TrimSpace(input.Image) == "" {
		return fmt.Errorf("image is required")
	}

	return nil
}

const swaggerDocument = `{
  "openapi": "3.0.3",
  "info": {
    "title": "Series REST API",
    "version": "1.0.0",
    "description": "CRUD API for managing series with pagination, search, and sorting."
  },
  "servers": [
    {
      "url": "http://localhost:8080"
    }
  ],
  "paths": {
    "/series": {
      "get": {
        "summary": "List all series",
        "parameters": [
          {
            "name": "page",
            "in": "query",
            "schema": {
              "type": "integer",
              "default": 1
            }
          },
          {
            "name": "limit",
            "in": "query",
            "schema": {
              "type": "integer",
              "default": 5
            }
          },
          {
            "name": "q",
            "in": "query",
            "schema": {
              "type": "string"
            }
          },
          {
            "name": "sort",
            "in": "query",
            "schema": {
              "type": "string",
              "enum": ["id", "title", "episodes"],
              "default": "id"
            }
          },
          {
            "name": "order",
            "in": "query",
            "schema": {
              "type": "string",
              "enum": ["asc", "desc"],
              "default": "asc"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Series list",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/SeriesListResponse"
                }
              }
            }
          },
          "400": {
            "description": "Invalid query",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          }
        }
      },
      "post": {
        "summary": "Create a series",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/SeriesInput"
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Created",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Series"
                }
              }
            }
          },
          "400": {
            "description": "Invalid data",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          }
        }
      }
    },
    "/series/{id}": {
      "get": {
        "summary": "Get a series by ID",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": {
              "type": "integer"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Series found",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Series"
                }
              }
            }
          },
          "404": {
            "description": "Not found",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          }
        }
      },
      "put": {
        "summary": "Update a series",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": {
              "type": "integer"
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/SeriesInput"
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Updated",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Series"
                }
              }
            }
          },
          "400": {
            "description": "Invalid data",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          },
          "404": {
            "description": "Not found",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          }
        }
      },
      "delete": {
        "summary": "Delete a series",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": {
              "type": "integer"
            }
          }
        ],
        "responses": {
          "204": {
            "description": "Deleted"
          },
          "404": {
            "description": "Not found",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "Series": {
        "type": "object",
        "required": ["id", "title", "description", "episodes", "image"],
        "properties": {
          "id": {
            "type": "integer"
          },
          "title": {
            "type": "string"
          },
          "description": {
            "type": "string"
          },
          "episodes": {
            "type": "integer"
          },
          "image": {
            "type": "string"
          }
        }
      },
      "SeriesInput": {
        "type": "object",
        "required": ["title", "description", "episodes", "image"],
        "properties": {
          "title": {
            "type": "string"
          },
          "description": {
            "type": "string"
          },
          "episodes": {
            "type": "integer",
            "minimum": 1
          },
          "image": {
            "type": "string"
          }
        }
      },
      "SeriesListResponse": {
        "type": "object",
        "properties": {
          "data": {
            "type": "array",
            "items": {
              "$ref": "#/components/schemas/Series"
            }
          },
          "page": {
            "type": "integer"
          },
          "limit": {
            "type": "integer"
          },
          "total": {
            "type": "integer"
          },
          "total_pages": {
            "type": "integer"
          }
        }
      },
      "ErrorResponse": {
        "type": "object",
        "properties": {
          "error": {
            "type": "string"
          }
        }
      }
    }
  }
}`
