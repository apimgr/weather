const chalk = require('chalk');
const weatherService = require('../services/weatherService');
const unitConverter = require('./unitConverter');

// Force colors to be enabled
chalk.level = 3; // Support for 256 colors

class WttrRenderer {
  constructor() {
    this.width = 120;
  }

  renderWttrStyle(locationData, currentWeather, forecast, units = 'imperial', hideFooter = false, formatInfo = null) {
    const lines = [];
    
    // Header (skip if quiet mode)
    if (!formatInfo || !formatInfo.quiet) {
      lines.push(this.renderHeader(locationData));
      lines.push('');
    }
    
    // Current weather with ASCII art
    lines.push(...this.renderCurrentWeatherArt(currentWeather, units));
    
    // Add forecast based on format
    if (formatInfo && formatInfo.days > 0) {
      lines.push('');
      lines.push(''); // Extra spacing before forecast table
      
      // Render forecast table with specified number of days
      const forecastDays = forecast.days.slice(0, formatInfo.days);
      const modifiedForecast = { ...forecast, days: forecastDays };
      lines.push(...this.renderForecastTable(modifiedForecast, units));
    } else if (!formatInfo || formatInfo.days === undefined) {
      // Default: show full forecast
      lines.push('');
      lines.push(''); // Extra spacing before forecast table
      lines.push(...this.renderForecastTable(forecast, units));
    }
    
    if (!hideFooter) {
      lines.push('');
      lines.push(this.renderFooter());
    }
    
    // Add two newlines at the end for proper terminal spacing
    return lines.join('\n') + '\n\n';
  }

  renderHeader(location) {
    const fullLocation = this.getFullLocationName(location);
    return chalk.hex('#f1fa8c').bold(`Weather report: ${fullLocation.toLowerCase()}`);
  }

  getFullLocationName(location) {
    // Use the pre-built full name from geocoding API if available
    if (location.fullName) {
      return location.fullName;
    }
    
    // Fallback: build full name
    return `${location.name}, ${location.country}`;
  }

  renderCurrentWeatherArt(weather, units) {
    const temp = Math.round(weather.temperature);
    const feelsLike = Math.round(weather.feelsLike);
    const tempUnit = unitConverter.getTemperatureUnit(units);
    const speedUnit = unitConverter.getSpeedUnit(units);
    const precipUnit = unitConverter.getPrecipitationUnit(units);
    const windSpeed = Math.round(weather.windSpeed);
    const description = weatherService.getWeatherDescription(weather.weatherCode);
    const precipitation = units === 'imperial' ? weather.precipitation.toFixed(1) : weather.precipitation.toFixed(1);
    
    const art = this.getColoredWeatherArt(weather.weatherCode, weather.isDay);
    const windDir = this.getWindArrow(weather.windDirection);
    
    // Create current weather display with Dracula colors
    return [
      `${art.line1}     ${chalk.hex('#8be9fd')(description)}`,
      `${art.line2}     ${chalk.hex('#f1fa8c')(`+${temp}(${feelsLike}) ${tempUnit}`)}`,
      `${art.line3}     ${chalk.hex('#50fa7b')(`${windDir} ${windSpeed} ${speedUnit}`)}`,
      `${art.line4}     ${chalk.hex('#bd93f9')(`${Math.round(weather.pressure * (units === 'imperial' ? 0.02953 : 1))} ${units === 'imperial' ? 'inHg' : 'hPa'}`)}`,
      `${art.line5}     ${chalk.hex('#ff79c6')(`${precipitation} ${precipUnit}`)}`
    ];
  }

  renderForecastTable(forecast, units) {
    const lines = [];
    const days = forecast.days.slice(0, 3);
    
    if (days.length === 0) return ['No forecast data available'];
    
    // Get first day for header
    const firstDay = new Date(days[0].date);
    const dayHeader = firstDay.toLocaleDateString('en-US', { 
      weekday: 'short',
      day: 'numeric', 
      month: 'short'
    });
    
    // Calculate dynamic spacing based on date length for perfect alignment
    const baseTableWidth = 123; // Base table without day header: 30+1+30+1+30+1+30 = 123
    const dayHeaderLength = dayHeader.length;
    const dayBoxTotalWidth = dayHeaderLength + 4; // Add padding: "  " + date + "  "
    const totalWidth = Math.max(baseTableWidth, dayBoxTotalWidth + 20); // Ensure minimum spacing
    
    // Top day header box - mathematically centered
    const dayBoxPadding = Math.floor((totalWidth - 13) / 2);
    lines.push(' '.repeat(dayBoxPadding) + chalk.hex('#bd93f9')('┌─────────────┐') + ' '.repeat(totalWidth - dayBoxPadding - 13));
    
    // Calculate left and right sections based on date length
    const leftSectionLength = 55; // Fixed: "┌──────────────────────────────┬───────────────────────┤  "
    const rightSectionLength = 55; // Fixed: " ├───────────────────────┬──────────────────────────────┐"
    const dateSpaceNeeded = dayHeaderLength + 2; // Space for date + padding
    
    // Adjust spacing around date to maintain table structure
    const leftDateSpace = Math.floor((dateSpaceNeeded - dayHeaderLength) / 2);
    const rightDateSpace = dateSpaceNeeded - dayHeaderLength - leftDateSpace;
    
    lines.push(
      chalk.hex('#bd93f9')('┌──────────────────────────────┬───────────────────────┤') +
      ' '.repeat(leftDateSpace + 1) + chalk.hex('#ff79c6')(dayHeader) + ' '.repeat(rightDateSpace + 1) +
      chalk.hex('#bd93f9')('├───────────────────────┬──────────────────────────────┐')
    );
    
    // Time period headers - fixed positioning
    lines.push(
      chalk.hex('#bd93f9')('│') + chalk.hex('#ffb86c')('            Morning           ') +
      chalk.hex('#bd93f9')('│') + chalk.hex('#ffb86c')('             Noon      ') +
      chalk.hex('#bd93f9')('└──────┬──────┘') + chalk.hex('#ffb86c')('     Evening           ') +
      chalk.hex('#bd93f9')('│') + chalk.hex('#ffb86c')('             Night            ') +
      chalk.hex('#bd93f9')('│')
    );
    
    // Data separator
    lines.push(chalk.hex('#bd93f9')('├──────────────────────────────┼──────────────────────────────┼──────────────────────────────┼──────────────────────────────┤'));
    
    // Add each day's forecast
    days.forEach(day => {
      const periods = this.generateDayPeriods(day, units);
      
      // Each period gets 5 lines of ASCII art + data with colors
      for (let lineIndex = 0; lineIndex < 5; lineIndex++) {
        const mornLine = this.padToWidth(periods.morning[lineIndex], 30);
        const noonLine = this.padToWidth(periods.noon[lineIndex], 30);
        const evenLine = this.padToWidth(periods.evening[lineIndex], 30);
        const nightLine = this.padToWidth(periods.night[lineIndex], 30);
        
        const line = chalk.hex('#bd93f9')('│') + mornLine + chalk.hex('#bd93f9')('│') + noonLine + chalk.hex('#bd93f9')('│') + evenLine + chalk.hex('#bd93f9')('│') + nightLine + chalk.hex('#bd93f9')('│');
        lines.push(line);
      }
      
      // Add separator between days (except last)
      if (day !== days[days.length - 1]) {
        lines.push(chalk.hex('#bd93f9')('├──────────────────────────────┼──────────────────────────────┼──────────────────────────────┼──────────────────────────────┤'));
        lines.push(''); // Add blank line between days for better readability
      }
    });
    
    lines.push(chalk.hex('#bd93f9')('└──────────────────────────────┴──────────────────────────────┴──────────────────────────────┴──────────────────────────────┘'));
    
    return lines;
  }

  generateDayPeriods(day, units) {
    const tempUnit = unitConverter.getTemperatureUnit(units);
    const speedUnit = unitConverter.getSpeedUnit(units);
    const precipUnit = unitConverter.getPrecipitationUnit(units);
    
    // Generate temperatures for different periods
    const tempLow = Math.round(day.tempMin);
    const tempHigh = Math.round(day.tempMax);
    const tempMorn = Math.round(tempLow + (tempHigh - tempLow) * 0.3);
    const tempNoon = tempHigh;
    const tempEven = Math.round(tempLow + (tempHigh - tempLow) * 0.7);
    const tempNight = tempLow;
    
    const windSpeed = Math.round(day.windSpeedMax);
    const windDir = this.getWindArrow(day.windDirection);
    const precipitation = units === 'imperial' ? day.precipitation.toFixed(1) : day.precipitation.toFixed(1);
    const precipProb = day.precipitationProbability || 0;
    
    const art = this.getWeatherArt(day.weatherCode, true); // Day periods
    const description = weatherService.getWeatherDescription(day.weatherCode);
    
    // Create properly aligned content - each line exactly fits the column
    const createPeriodLines = (temp, feels, windMod = 1) => {
      const tempLine = `+${temp}(${feels}) ${tempUnit}`;
      const windLine = `${windDir} ${Math.max(1, Math.round(windSpeed * windMod))} ${speedUnit}`;
      const precipLine = `${precipitation} ${precipUnit} | ${precipProb}%`;
      
      return [
        this.centerInWidth(chalk.hex('#8be9fd')(description.substring(0, 12)), 30),
        this.centerInWidth(chalk.hex('#f1fa8c')(tempLine), 30),
        this.centerInWidth(chalk.hex('#50fa7b')(windLine), 30),
        this.centerInWidth(chalk.hex('#bd93f9')('6 mi'), 30),
        this.centerInWidth(chalk.hex('#ff79c6')(precipitation + ' ' + precipUnit) + ' | ' + chalk.hex('#ffb86c')(precipProb + '%'), 30)
      ];
    };
    
    return {
      morning: createPeriodLines(tempMorn, tempMorn + 2, 0.8),
      noon: createPeriodLines(tempNoon, tempNoon + 3, 1.0),
      evening: createPeriodLines(tempEven, tempEven + 2, 1.2),
      night: createPeriodLines(tempNight, tempNight + 1, 0.6)
    };
  }

  getColoredWeatherArt(weatherCode, isDay = true) {
    const baseArt = this.getWeatherArt(weatherCode, isDay);
    
    // Apply Dracula theme colors to ASCII art
    const artColor = this.getWeatherColor(weatherCode, isDay);
    
    return {
      line1: chalk.hex(artColor)(baseArt.line1),
      line2: chalk.hex(artColor)(baseArt.line2),
      line3: chalk.hex(artColor)(baseArt.line3),
      line4: chalk.hex(artColor)(baseArt.line4),
      line5: chalk.hex(artColor)(baseArt.line5)
    };
  }

  getWeatherColor(weatherCode, isDay) {
    // Dracula theme colors for different weather conditions
    if (weatherCode === 0) return isDay ? '#f1fa8c' : '#bd93f9'; // Clear: yellow/purple
    if (weatherCode <= 3) return '#6272a4'; // Cloudy: comment gray
    if (weatherCode >= 45 && weatherCode <= 48) return '#f8f8f2'; // Fog: foreground
    if (weatherCode >= 51 && weatherCode <= 67) return '#8be9fd'; // Rain: cyan
    if (weatherCode >= 71 && weatherCode <= 86) return '#f8f8f2'; // Snow: white
    if (weatherCode >= 95) return '#ff5555'; // Storms: red
    
    return '#f8f8f2'; // Default: foreground
  }

  getWeatherArt(weatherCode, isDay = true) {
    const arts = {
      // Clear sky
      0: isDay ? {
        line1: '    \\   /    ',
        line2: '     .-.     ', 
        line3: '  ― (   ) ―  ',
        line4: '     `-\'     ',
        line5: '    /   \\    '
      } : {
        line1: '             ',
        line2: '     .--.    ',
        line3: '  .-(    ).  ',
        line4: ' (___.__)_)  ',
        line5: '             '
      },
      
      // Mainly clear
      1: {
        line1: '   \\  /      ',
        line2: ' _ /\"\".\-.    ',
        line3: '   \\_(   ).  ',
        line4: '   /(_(__)   ',
        line5: '             '
      },
      
      // Partly cloudy / Overcast
      2: {
        line1: '             ',
        line2: '     .--.    ',
        line3: '  .-(    ).  ',
        line4: ' (___.__)_)  ',
        line5: '             '
      },
      
      // Rain
      61: {
        line1: '     .-.     ',
        line2: '    (   ).   ',
        line3: '   (___(__)  ',
        line4: '    ‚\'‚\'‚\'‚\' ',
        line5: '    ‚\'‚\'‚\'‚\' '
      }
    };
    
    return arts[weatherCode] || arts[2]; // Default to cloudy
  }

  getWindArrow(degrees) {
    const arrows = ['↓', '↙', '←', '↖', '↑', '↗', '→', '↘'];
    const index = Math.round(degrees / 45) % 8;
    return arrows[index];
  }

  renderCurrentOnly(locationData, currentWeather, units, hideFooter = false) {
    const lines = [];
    
    // Header
    lines.push(this.renderHeader(locationData));
    lines.push('');
    
    // Current weather with ASCII art only (no forecast)
    lines.push(...this.renderCurrentWeatherArt(currentWeather, units));
    
    if (!hideFooter) {
      lines.push('');
      lines.push(this.renderFooter());
    }
    
    return lines.join('\n') + '\n\n';
  }

  renderFooter() {
    return chalk.hex('#6272a4')('Console Weather Service • Free weather data from Open-Meteo.com');
  }

  padToWidth(text, width) {
    // Remove ANSI color codes to get the actual text length
    const cleanText = text.replace(/\u001b\[[0-9;]*m/g, '');
    const actualLength = cleanText.length;
    const padding = Math.max(0, width - actualLength);
    return text + ' '.repeat(padding);
  }

  stripAnsiCodes(text) {
    return text.replace(/\u001b\[[0-9;]*m/g, '');
  }

  centerInWidth(text, width) {
    const cleanText = this.stripAnsiCodes(text);
    const textLength = cleanText.length;
    const totalPadding = Math.max(0, width - textLength);
    const leftPadding = Math.floor(totalPadding / 2);
    const rightPadding = totalPadding - leftPadding;
    
    return ' '.repeat(leftPadding) + text + ' '.repeat(rightPadding);
  }
}

module.exports = new WttrRenderer();