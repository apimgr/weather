class UserAgentDetector {
  isPlainTextAgent(userAgent) {
    if (!userAgent) return false;
    
    const plainTextAgents = [
      'curl',
      'httpie', 
      'lwp-request',
      'wget',
      'python-httpx',
      'python-requests',
      'openbsd ftp',
      'powershell',
      'fetch',
      'aiohttp',
      'http_get',
      'xh',
      'nushell',
      'zig',
      'node'
    ];
    
    const ua = userAgent.toLowerCase();
    return plainTextAgents.some(agent => ua.includes(agent));
  }

  isBrowser(userAgent) {
    if (!userAgent) return false;
    
    const ua = userAgent.toLowerCase();
    
    // Detect browsers more broadly
    const browserSignatures = [
      'chrome', 'firefox', 'safari', 'edge', 'opera', 'brave',
      'webkit', 'gecko', 'trident', 'blink'
    ];
    
    // Must contain mozilla AND one of the browser signatures
    // This excludes command line tools that might include mozilla
    return ua.includes('mozilla') && 
           browserSignatures.some(sig => ua.includes(sig)) &&
           !this.isPlainTextAgent(userAgent);
  }

  getPreferredFormat(userAgent, requestedFormat) {
    // If format is explicitly requested, honor it
    if (requestedFormat) {
      return requestedFormat;
    }

    // All clients should get the full ASCII art format by default
    // Only use simple format when explicitly requested
    return 'full';
  }

  shouldUseColors(userAgent) {
    if (!userAgent) return false;
    
    // These agents typically support ANSI colors in terminal
    const colorSupportedAgents = ['curl', 'httpie', 'wget'];
    const ua = userAgent.toLowerCase();
    
    return colorSupportedAgents.some(agent => ua.includes(agent));
  }

  getClientInfo(req) {
    const userAgent = req.headers['user-agent'] || '';
    
    return {
      userAgent,
      isPlainText: this.isPlainTextAgent(userAgent),
      isBrowser: this.isBrowser(userAgent),
      supportsColors: this.shouldUseColors(userAgent),
      preferredFormat: this.getPreferredFormat(userAgent, req.query.format)
    };
  }
}

module.exports = new UserAgentDetector();