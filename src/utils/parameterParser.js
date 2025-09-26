class ParameterParser {
  parseWttrParams(query) {
    const params = {
      format: 'full',    // Default format
      units: 'auto',     // Auto-detect units
      lang: 'en',        // Language
      narrow: false,     // Narrow format
      quiet: false,      // Quiet mode (no location)
      oneShot: false,    // One-line format
      transparency: false, // PNG transparency
      followRedirects: true,
      showFooter: true
    };

    // Format options - exact wttr.in format numbers
    if (query.format !== undefined) {
      switch(query.format) {
        case '0': params.format = 0; break;  // Special format 0
        case '1': params.format = 1; break;  // Icon + temp
        case '2': params.format = 2; break;  // Icon + temp + wind
        case '3': params.format = 3; break;  // Location + icon + temp
        case '4': params.format = 4; break;  // Location + icon + temp + wind
        case 'simple': params.format = 'simple'; break;
        default: params.format = query.format;
      }
    }

    // Unit options
    if (query.u !== undefined) {
      params.units = 'imperial'; // u = USCS/Imperial
    } else if (query.m !== undefined) {
      params.units = 'metric'; // m = Metric
    } else if (query.M !== undefined) {
      params.units = 'metric'; // M = SI (metric)
    } else if (query.units) {
      params.units = query.units;
    }

    // View options
    if (query.n !== undefined) {
      params.narrow = true; // Narrow format
    }
    if (query.q !== undefined) {
      params.quiet = true; // Quiet mode
    }
    if (query.Q !== undefined) {
      params.quiet = true; // Super quiet
    }
    if (query.T !== undefined) {
      params.transparency = true; // PNG transparency
    }

    // Footer control
    if (query.F !== undefined) {
      params.showFooter = false;
    }

    // Language
    if (query.lang) {
      params.lang = query.lang;
    }

    // One-line format for status bars
    if (query['0'] !== undefined) {
      params.format = 'one-line';
      params.oneShot = true;
    }

    return params;
  }

  getWeatherFormat(params) {
    switch(params.format) {
      case 0:
        return { format: 0, currentOnly: true }; // Current weather only (no forecast)
      case 1:
        return { format: 1, oneLine: true }; // Just: 🌦 +11⁰C
      case 2:
        return { format: 2, oneLine: true, detailed: true }; // Just: 🌦   🌡️+11°C 🌬️↓4km/h
      case 3:
        return { format: 3, oneLine: true, showLocation: true }; // Location: 🌦 +11⁰C
      case 4:
        return { format: 4, oneLine: true, showLocation: true, detailed: true }; // Location: 🌦   🌡️+11°C 🌬️↓4km/h
      case 'simple':
        return { simple: true };
      case 'one-line':
        return { oneLine: true };
      default:
        return { full: true }; // Default full ASCII art format
    }
  }

  shouldShowLocation(params) {
    return !params.quiet;
  }

  getUnitsFromParams(params, detectedUnits) {
    if (params.units === 'auto') {
      return detectedUnits || 'imperial';
    }
    return params.units;
  }
}

module.exports = new ParameterParser();