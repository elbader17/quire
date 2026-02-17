package quire

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type sheetsClient struct {
	srv           *sheets.Service
	spreadsheetID string
}

func newSheetsClient(credentials []byte, spreadsheetID string) (*sheetsClient, error) {
	ctx := context.Background()

	srv, err := sheets.NewService(ctx, option.WithCredentialsJSON(credentials))
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	return &sheetsClient{
		srv:           srv,
		spreadsheetID: spreadsheetID,
	}, nil
}

func (c *sheetsClient) Read(ctx context.Context, range_ string) ([][]interface{}, error) {
	resp, err := c.srv.Spreadsheets.Values.Get(c.spreadsheetID, range_).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read range %s: %w", range_, err)
	}
	return resp.Values, nil
}

func (c *sheetsClient) Write(ctx context.Context, range_ string, values [][]interface{}) error {
	valueRange := &sheets.ValueRange{
		Values: values,
	}

	_, err := c.srv.Spreadsheets.Values.Update(c.spreadsheetID, range_, valueRange).
		ValueInputOption("RAW").
		Context(ctx).
		Do()

	if err != nil {
		return fmt.Errorf("failed to write to range %s: %w", range_, err)
	}
	return nil
}

func (c *sheetsClient) Append(ctx context.Context, range_ string, values [][]interface{}) error {
	valueRange := &sheets.ValueRange{
		Values: values,
	}

	_, err := c.srv.Spreadsheets.Values.Append(c.spreadsheetID, range_, valueRange).
		ValueInputOption("RAW").
		InsertDataOption("INSERT_ROWS").
		Context(ctx).
		Do()

	if err != nil {
		return fmt.Errorf("failed to append to range %s: %w", range_, err)
	}
	return nil
}

func (c *sheetsClient) Clear(ctx context.Context, range_ string) error {
	_, err := c.srv.Spreadsheets.Values.Clear(c.spreadsheetID, range_, &sheets.ClearValuesRequest{}).
		Context(ctx).
		Do()

	if err != nil {
		return fmt.Errorf("failed to clear range %s: %w", range_, err)
	}
	return nil
}
