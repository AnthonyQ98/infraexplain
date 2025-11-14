import { useState, useEffect } from 'react'
import './App.css'

// Use relative path to leverage Vite proxy in dev, or full URL in production
const API_BASE_URL = import.meta.env.VITE_API_URL || ''

function App() {
  const [terraformCode, setTerraformCode] = useState('')
  const [explanation, setExplanation] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleExplain = async () => {
    if (!terraformCode.trim()) {
      setError('Please enter some Terraform code to explain.')
      return
    }

    setLoading(true)
    setError('')
    setExplanation('')

    try {
      const response = await fetch(`${API_BASE_URL}/explain`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          text_content: terraformCode,
        }),
      })

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(errorText || `Server error: ${response.status}`)
      }

      const data = await response.json()
      setExplanation(data.summary || 'No explanation available.')
    } catch (err) {
      // Don't log sensitive error details to console
      let errorMsg = err.message
      if (err.message.includes('Failed to fetch') || err.message.includes('NetworkError')) {
        errorMsg = 'Unable to connect to the backend server. Make sure it\'s running on port 8080.'
      }
      setError(`Failed to get explanation: ${errorMsg}`)
    } finally {
      setLoading(false)
    }
  }

  const handleClear = () => {
    setTerraformCode('')
    setExplanation('')
    setError('')
  }

  const handleKeyDown = (e) => {
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
      handleExplain()
    }
  }

  const handlePaste = (e) => {
    // Get pasted content from the paste event
    const pastedText = e.clipboardData?.getData('text') || ''
    
    // Check if pasted text contains literal escaped sequences (like \n, \t, etc.)
    // This handles cases where someone pastes escaped JSON strings
    if (pastedText && (pastedText.includes('\\n') || pastedText.includes('\\t') || pastedText.includes('\\"'))) {
      e.preventDefault()
      // Convert escaped sequences to actual characters
      const cleanedText = pastedText
        .replace(/\\n/g, '\n')      // \n -> newline
        .replace(/\\t/g, '\t')      // \t -> tab
        .replace(/\\r/g, '\r')      // \r -> carriage return
        .replace(/\\"/g, '"')        // \" -> "
        .replace(/\\\\/g, '\\')      // \\ -> \ (but only if not followed by n, t, r, or ")
      setTerraformCode(cleanedText)
    }
    // Otherwise, let default paste behavior handle it normally
  }

  const handleChange = (e) => {
    let value = e.target.value
    
    // Clean up escaped sequences if they appear (handles manual typing or paste that didn't trigger handlePaste)
    // Only do this if we see escaped sequences but few actual newlines
    const hasEscapedNewlines = value.includes('\\n')
    const actualNewlineCount = (value.match(/\n/g) || []).length
    
    if (hasEscapedNewlines && actualNewlineCount < 3) {
      // Likely pasted escaped text, clean it up
      value = value
        .replace(/\\n/g, '\n')
        .replace(/\\t/g, '\t')
        .replace(/\\r/g, '\r')
        .replace(/\\"/g, '"')
    }
    
    setTerraformCode(value)
  }

  // Clear all data on component unmount for privacy
  useEffect(() => {
    return () => {
      setTerraformCode('')
      setExplanation('')
      setError('')
    }
  }, [])

  return (
    <div className="app">
      <div className="container">
        <header>
          <h1>InfraExplain</h1>
          <p className="subtitle">Understand your Terraform infrastructure code</p>
          <div className="privacy-badge">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M8 1C6.34 1 5 2.34 5 4V6H4C3.45 6 3 6.45 3 7V13C3 13.55 3.45 14 4 14H12C12.55 14 13 13.55 13 13V7C13 6.45 12.55 6 12 6H11V4C11 2.34 9.66 1 8 1ZM9 6H7V4C7 3.45 7.45 3 8 3C8.55 3 9 3.45 9 4V6Z" fill="currentColor"/>
            </svg>
            <span>No data stored • All processing is ephemeral</span>
          </div>
        </header>

        <main>
          <div className="input-section">
            <label htmlFor="terraform-code">Terraform Code</label>
            <textarea
              id="terraform-code"
              value={terraformCode}
              onChange={handleChange}
              onKeyDown={handleKeyDown}
              onPaste={handlePaste}
              placeholder="Paste your Terraform code here..."
              rows={12}
              disabled={loading}
            />
            <button
              className="btn-primary"
              onClick={handleExplain}
              disabled={loading || !terraformCode.trim()}
            >
              {loading ? (
                <>
                  <span className="btn-loader">⏳</span>
                  <span>Processing...</span>
                </>
              ) : (
                'Explain'
              )}
            </button>
          </div>

          {error && (
            <div className="error-message" role="alert">
              {error}
            </div>
          )}

          {explanation && (
            <div className="result-section">
              <label>Explanation</label>
              <div className="explanation-box">
                {explanation.split('\n').map((line, index) => {
                  // Handle markdown-style bold (**text**)
                  const parts = line.split(/(\*\*.*?\*\*)/g)
                  return (
                    <div key={index} style={{ marginBottom: '0.5em' }}>
                      {parts.map((part, partIndex) => {
                        if (part.startsWith('**') && part.endsWith('**')) {
                          return (
                            <strong key={partIndex}>
                              {part.slice(2, -2)}
                            </strong>
                          )
                        }
                        return <span key={partIndex}>{part}</span>
                      })}
                    </div>
                  )
                })}
              </div>
              <button className="btn-secondary" onClick={handleClear}>
                Clear All
              </button>
            </div>
          )}
        </main>
      </div>
    </div>
  )
}

export default App

