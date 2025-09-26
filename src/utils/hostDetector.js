class HostDetector {
  getHostInfo(req) {
    const protocol = req.headers['x-forwarded-proto'] || 
                     req.protocol || 
                     (req.secure ? 'https' : 'http');
                     
    const host = req.headers['x-forwarded-host'] || 
                 req.headers['host'] || 
                 'localhost:3000';
                 
    const fullHost = `${protocol}://${host}`;
    
    return {
      protocol,
      host,
      fullHost,
      port: req.app.get('port') || process.env.PORT || 3000
    };
  }

  generateExampleUrls(req) {
    const { fullHost } = this.getHostInfo(req);
    
    return {
      console: {
        current: `curl ${fullHost}/`,
        location: `curl ${fullHost}/London`,
        simple: `curl ${fullHost}/Paris?format=simple`,
        units: `curl ${fullHost}/Tokyo?units=metric`
      },
      api: {
        weather: `curl ${fullHost}/api/v1/weather?location=London`,
        current: `curl ${fullHost}/api/v1/current?location=Paris`,
        forecast: `curl ${fullHost}/api/v1/forecast?location=Tokyo&days=5`,
        search: `curl ${fullHost}/api/v1/search?q=New+York`
      }
    };
  }

  getBaseUrl(req) {
    return this.getHostInfo(req).fullHost;
  }
}

module.exports = new HostDetector();