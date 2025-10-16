# MIT License

Copyright (c) 2024 Console Weather Service Contributors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

---

# Third-Party Licenses and Attributions

## Go Dependencies

### Gin Web Framework
- **License**: MIT
- **Copyright**: 2014 Manuel Martínez-Almeida and Gin contributors
- **Website**: https://gin-gonic.com/
- **Repository**: https://github.com/gin-gonic/gin

### Gin Contrib - CORS
- **License**: MIT
- **Copyright**: 2016 Gin-Gonic contributors
- **Repository**: https://github.com/gin-contrib/cors

### Gin Contrib - Secure
- **License**: MIT
- **Copyright**: 2015 Gin-Gonic contributors
- **Repository**: https://github.com/gin-contrib/secure

## Data Sources

### Open-Meteo Weather API
- **License**: CC BY 4.0 (Creative Commons Attribution 4.0 International)
- **Data Provider**: Open-Meteo.com
- **Website**: https://open-meteo.com/
- **Terms**: https://open-meteo.com/en/terms
- **Attribution**: Weather data provided by Open-Meteo.com
- **Usage**: Free for non-commercial and commercial use with attribution

### OpenStreetMap Nominatim (Reverse Geocoding)
- **License**: ODbL (Open Database License)
- **Copyright**: OpenStreetMap contributors
- **Website**: https://nominatim.openstreetmap.org/
- **Terms**: https://operations.osmfoundation.org/policies/nominatim/
- **Attribution**: © OpenStreetMap contributors
- **Usage**: Free with attribution and fair use policy

### Cities Database
- **Source**: OpenWeatherMap City List
- **Repository**: https://github.com/apimgr/citylist
- **License**: CC BY-SA 4.0
- **Records**: 209,579 cities worldwide
- **Attribution**: City data compiled from OpenWeatherMap and OpenStreetMap

### Countries Database
- **Source**: REST Countries API
- **Repository**: https://github.com/apimgr/countries
- **License**: MPL 2.0 (Mozilla Public License 2.0)
- **Records**: 247 countries with timezones and metadata
- **Attribution**: Country data from various open sources

## Design and Theme

### Dracula Theme
- **License**: MIT
- **Copyright**: 2016 Dracula Theme contributors
- **Website**: https://draculatheme.com/
- **Repository**: https://github.com/dracula/dracula-theme
- **Colors**: Color palette and design inspiration from Dracula Theme

## Inspiration

### wttr.in
- **License**: Apache-2.0
- **Author**: Igor Chubin
- **Website**: https://wttr.in
- **Repository**: https://github.com/chubin/wttr.in
- **Inspiration**: Format parameter design and ASCII art concepts

## Fonts and Typography

### System Monospace Fonts
The service uses system monospace fonts for terminal display:
- **Monaco** (macOS) - Apple Inc.
- **Menlo** (macOS) - Apple Inc.
- **Ubuntu Mono** (Linux) - Canonical Ltd., Ubuntu Font Licence
- **Consolas** (Windows) - Microsoft Corporation
- **Courier New** (fallback) - Various

## Container and Deployment

### Docker
- **License**: Apache-2.0
- **Copyright**: 2013-2024 Docker, Inc.
- **Website**: https://www.docker.com/

### Alpine Linux (Docker Base Image)
- **License**: Various open source licenses
- **Website**: https://alpinelinux.org/
- **Usage**: Minimal container base image

## Build Tools

### GitHub Actions
- **License**: MIT
- **Copyright**: 2019-2024 GitHub, Inc.
- **Website**: https://github.com/features/actions
- **Usage**: CI/CD automation for multi-platform builds

## Attribution Requirements

When using this software, please provide attribution to:

1. **Console Weather Service** - https://github.com/apimgr/weather
2. **Open-Meteo** - Weather data provided by Open-Meteo.com
3. **OpenStreetMap** - Geocoding data © OpenStreetMap contributors

### Example Attribution

```
Weather data provided by Open-Meteo.com
Geocoding data © OpenStreetMap contributors
Powered by Console Weather Service
```

## Data Usage and Privacy

### Weather Data
- Weather data is fetched from Open-Meteo.com in real-time
- No personal data is collected or stored
- Cached for 10 minutes to reduce API calls

### Location Data
- IP-based location detection uses HTTP headers
- No IP addresses are logged or stored
- GPS coordinates are processed server-side only
- No tracking or analytics

### Cookies and Storage
- No cookies are set by this service
- No user data is stored
- No third-party tracking scripts

## Disclaimer

THIS SOFTWARE AND RELATED DATA ARE PROVIDED "AS IS" WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, AND NONINFRINGEMENT.

Weather data accuracy is not guaranteed and should not be used for critical decision-making or emergency planning. Always consult official weather services for important weather information.

Location data accuracy may vary depending on the data source and should not be used for navigation or precise positioning.

## Contributing

By contributing to this project, you agree that:
- Your contributions will be licensed under the same MIT License
- You have the right to submit the contributions
- Your contributions are your original work or properly attributed

## Open Source Commitment

This project is committed to:
- 🔓 Open source development
- 🆓 Free access to weather information
- 🌍 Global accessibility
- 🤝 Community contributions
- 📖 Transparent operations

## Acknowledgments

Special thanks to:
- The Go programming language community
- Open-Meteo for providing free weather data
- OpenStreetMap contributors for geocoding data
- wttr.in for inspiration and format design
- Dracula Theme for the beautiful color scheme
- All contributors who helped improve this project

## Contact and Support

- **Repository**: https://github.com/apimgr/weather
- **Issues**: https://github.com/apimgr/weather/issues
- **Discussions**: https://github.com/apimgr/weather/discussions
- **Live Demo**: http://wthr.top

## Updates

- **Last Updated**: 2024
- **License Version**: 1.0
- **Go Version**: 2.0.0

---

For the full list of dependencies and their licenses, see `go.mod` and run:
```bash
go list -m -json all
```
