// internal/common/mcp_client.go
package common

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// MCPClient represents a client for interacting with the MCP server
type MCPClient struct {
	client  *client.Client
	baseURL string
	ctx     context.Context
	cancel  context.CancelFunc
}

// MCPResponse represents the response from the MCP server
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an error from the MCP server
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// DatabaseInfo represents information about a vector database
type DatabaseInfo struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Collection    string `json:"collection"`
	DocumentCount int    `json:"document_count"`
}

// normalizeURL ensures the URL has a protocol prefix and MCP endpoint
func normalizeURL(url string) string {
	// If it already has a protocol, just ensure it has the /mcp endpoint
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		if !strings.HasSuffix(url, "/mcp") {
			return strings.TrimSuffix(url, "/") + "/mcp"
		}
		return url
	}

	// If it's just a hostname:port, add http:// prefix and /mcp endpoint
	if strings.Contains(url, ":") {
		return "http://" + strings.TrimSuffix(url, "/") + "/mcp"
	}

	// If it's just a hostname, add http:// prefix, default port, and /mcp endpoint
	return "http://" + url + ":8030/mcp"
}

// getMCPServerURI gets the MCP server URI from environment variable or command line flag
func GetMCPServerURI(cmdServerURI string) (string, error) {
	// Load .env file if it exists
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return "", fmt.Errorf("failed to load .env file: %w", err)
		}
	}

	var serverURI string
	if cmdServerURI != "" {
		serverURI = cmdServerURI
	} else if envURI := os.Getenv("MAESTRO_KNOWLEDGE_MCP_SERVER_URI"); envURI != "" {
		serverURI = envURI
	} else {
		serverURI = "localhost:8030" // Default
	}

	return normalizeURL(serverURI), nil
}

// NewMCPClient creates a new MCP client
func NewMCPClient(serverURI string) (*MCPClient, error) {
	// Create context with timeout - use shorter timeout for tests
	timeout := 30 * time.Second
	if os.Getenv("MAESTRO_K_TEST_MODE") == "true" {
		timeout = 5 * time.Second // Shorter timeout for tests
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	// Create MCP client using HTTP transport
	// The mark3labs/mcp-go library supports HTTP transport for connecting to existing servers
	mcpClient, err := client.NewStreamableHttpClient(serverURI)
	if err != nil {
		// Cancel context on error to prevent context leak
		cancel()

		// Provide user-friendly error messages for common connection issues
		errStr := err.Error()
		if strings.Contains(errStr, "connection refused") ||
			strings.Contains(errStr, "no such host") ||
			strings.Contains(errStr, "timeout") ||
			strings.Contains(errStr, "context deadline exceeded") ||
			strings.Contains(errStr, "network is unreachable") {
			return nil, fmt.Errorf("MCP server could not be reached at %s. Please ensure the server is running and accessible", serverURI)
		}
		return nil, fmt.Errorf("failed to create MCP client: %w", err)
	}

	return &MCPClient{
		client:  mcpClient,
		baseURL: serverURI,
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

// callMCPServer makes a call to the MCP server using the mark3labs/mcp-go library
func (c *MCPClient) CallMCPServer(method string, params interface{}) (*MCPResponse, error) {

	// Initialize the client if not already initialized
	if !c.client.IsInitialized() {
		initRequest := mcp.InitializeRequest{
			Request: mcp.Request{
				Method: "initialize",
			},
			Params: mcp.InitializeParams{
				ProtocolVersion: "2024-11-05",
				Capabilities:    mcp.ClientCapabilities{},
			},
		}

		_, err := c.client.Initialize(c.ctx, initRequest)
		if err != nil {
			// Provide user-friendly error messages for common connection issues
			errStr := err.Error()
			if strings.Contains(errStr, "connection refused") ||
				strings.Contains(errStr, "no such host") ||
				strings.Contains(errStr, "timeout") ||
				strings.Contains(errStr, "context deadline exceeded") ||
				strings.Contains(errStr, "network is unreachable") {
				return nil, fmt.Errorf("MCP server could not be reached at %s. Please ensure the server is running and accessible", c.baseURL)
			}
			return nil, fmt.Errorf("failed to initialize MCP client: %w", err)
		}
	}

	// Create the tool call request
	request := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: method,
		},
		Params: mcp.CallToolParams{
			Name:      method,
			Arguments: params,
		},
	}

	// Call the tool
	response, err := c.client.CallTool(c.ctx, request)
	if err != nil {
		// Provide user-friendly error messages for common connection issues
		errStr := err.Error()
		if strings.Contains(errStr, "connection refused") ||
			strings.Contains(errStr, "no such host") ||
			strings.Contains(errStr, "timeout") ||
			strings.Contains(errStr, "context deadline exceeded") ||
			strings.Contains(errStr, "network is unreachable") {
			return nil, fmt.Errorf("MCP server could not be reached at %s. Please ensure the server is running and accessible", c.baseURL)
		}
		return nil, fmt.Errorf("failed to call MCP tool %s: %w", method, err)
	}

	// Convert the response to our format
	result := &MCPResponse{
		JSONRPC: "2.0",
		ID:      1,
	}

	// Check if the response indicates an error
	if response != nil && len(response.Content) > 0 {
		// Try to get text content
		if textContent, ok := mcp.AsTextContent(response.Content[0]); ok {
			// Check if the content contains an error message
			contentText := textContent.Text
			if strings.Contains(contentText, "ValueError:") ||
				strings.Contains(contentText, "Error:") ||
				strings.Contains(contentText, "Exception:") ||
				strings.Contains(contentText, "Error calling tool") {
				// This is an error response
				result.Error = &MCPError{
					Code:    -1,
					Message: contentText,
				}
				return result, nil
			}

			// Try to parse as JSON first
			var jsonResult interface{}
			if err := json.Unmarshal([]byte(contentText), &jsonResult); err == nil {
				result.Result = jsonResult
			} else {
				// If not JSON, use as string
				result.Result = contentText
			}
		}
	}

	return result, nil
}

// Close closes the MCP client
func (c *MCPClient) Close() error {
	// Cancel the context to prevent context leaks
	if c.cancel != nil {
		c.cancel()
	}

	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// CreateVectorDatabase calls the create_vector_database_tool on the MCP server
func (c *MCPClient) CreateVectorDatabase(dbName, dbType, collectionName string) error {
	params := map[string]interface{}{
		"input": map[string]interface{}{
			"db_name":         dbName,
			"db_type":         dbType,
			"collection_name": collectionName,
		},
	}

	response, err := c.CallMCPServer("create_vector_database_tool", params)
	if err != nil {
		return err
	}

	// Check for error in response
	if response.Error != nil {
		return fmt.Errorf("MCP server error: %s", response.Error.Message)
	}

	// The response should be a success message
	if response.Result == nil {
		return fmt.Errorf("no response from MCP server (check server at %s)", c.baseURL)
	}

	return nil
}
