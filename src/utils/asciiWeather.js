const chalk = require('chalk');
const weatherService = require('../services/weatherService');
const unitConverter = require('./unitConverter');

class AsciiWeatherRenderer {
  constructor() {
    this.width = 120;
  }

  renderWeatherReport(locationData, currentWeather, forecast, units = 'imperial', hideFooter = false) {
    const lines = [];
    
    lines.push(this.renderHeader(locationData));
    lines.push('');
    lines.push(this.renderCurrent(currentWeather, units));
    lines.push('');
    lines.push(this.renderForecast(forecast, units));
    
    if (!hideFooter) {
      lines.push('');
      lines.push(this.renderFooter());
    }
    
    return lines.join('\n');
  }

  renderHeader(location) {
    const title = `Weather report: ${location.name}, ${location.country}`;
    const padding = Math.max(0, Math.floor((this.width - title.length) / 2));
    return chalk.bold.cyan(' '.repeat(padding) + title);
  }

  renderCurrent(weather, units = 'imperial') {
    const temp = Math.round(weather.temperature);
    const feelsLike = Math.round(weather.feelsLike);
    const description = weatherService.getWeatherDescription(weather.weatherCode);
    const icon = weatherService.getWeatherIcon(weather.weatherCode, weather.isDay);
    const tempUnit = unitConverter.getTemperatureUnit(units);
    const speedUnit = unitConverter.getSpeedUnit(units);
    const pressureUnit = unitConverter.getPressureUnit(units);
    const precipUnit = unitConverter.getPrecipitationUnit(units);
    
    const currentBlock = [
      chalk.yellow('┌─────────────────────────────┐'),
      chalk.yellow('│') + chalk.bold(`  Current Weather (${icon})      `) + chalk.yellow('│'),
      chalk.yellow('│') + `  ${description}${' '.repeat(27 - description.length)}` + chalk.yellow('│'),
      chalk.yellow('│') + chalk.bold.white(`  ${temp}${tempUnit} `) + chalk.gray(`(feels like ${feelsLike}${tempUnit})`) + ' '.repeat(Math.max(0, 14 - temp.toString().length - feelsLike.toString().length)) + chalk.yellow('│'),
      chalk.yellow('└─────────────────────────────┘')
    ];

    const pressure = units === 'imperial' ? weather.pressure.toFixed(2) : Math.round(weather.pressure);
    const windSpeed = Math.round(weather.windSpeed);
    const precipitation = units === 'imperial' ? weather.precipitation.toFixed(2) : weather.precipitation.toFixed(1);

    const detailsBlock = [
      chalk.blue('┌─────────────────────────────┐'),
      chalk.blue('│') + chalk.bold('  Weather Details            ') + chalk.blue('│'),
      chalk.blue('│') + `  Humidity: ${weather.humidity}%${' '.repeat(19 - weather.humidity.toString().length)}` + chalk.blue('│'),
      chalk.blue('│') + `  Pressure: ${pressure} ${pressureUnit}${' '.repeat(Math.max(0, 16 - pressure.toString().length - pressureUnit.length))}` + chalk.blue('│'),
      chalk.blue('│') + `  Wind: ${windSpeed} ${speedUnit} ${this.getWindDirection(weather.windDirection)}${' '.repeat(Math.max(0, 12 - windSpeed.toString().length - speedUnit.length))}` + chalk.blue('│'),
      chalk.blue('│') + `  Precipitation: ${precipitation} ${precipUnit}${' '.repeat(Math.max(0, 14 - precipitation.toString().length - precipUnit.length))}` + chalk.blue('│'),
      chalk.blue('│') + `  Cloud cover: ${weather.cloudCover}%${' '.repeat(15 - weather.cloudCover.toString().length)}` + chalk.blue('│'),
      chalk.blue('└─────────────────────────────┘')
    ];

    return this.combineBlocks([currentBlock, detailsBlock]);
  }

  renderForecast(forecast, units = 'imperial') {
    const lines = [chalk.bold.green('7-Day Forecast:')];
    const tempUnit = unitConverter.getTemperatureUnit(units).replace('°', '');
    const speedUnit = unitConverter.getSpeedUnit(units);
    
    lines.push('┌──────────┬────────────────┬─────────┬──────────┬────────────┐');
    lines.push(`│   Date   │   Condition    │ Temp ${tempUnit}${' '.repeat(4 - tempUnit.length)} │ Precip % │    Wind    │`);
    lines.push('├──────────┼────────────────┼─────────┼──────────┼────────────┤');

    forecast.days.slice(0, 7).forEach(day => {
      const date = new Date(day.date).toLocaleDateString('en-US', { 
        weekday: 'short', 
        month: 'short', 
        day: 'numeric' 
      });
      const condition = weatherService.getWeatherDescription(day.weatherCode);
      const icon = weatherService.getWeatherIcon(day.weatherCode);
      const tempRange = `${Math.round(day.tempMin)}/${Math.round(day.tempMax)}`;
      const precip = day.precipitationProbability || 0;
      const windSpeed = Math.round(day.windSpeedMax);
      const wind = `${windSpeed} ${speedUnit}`;
      
      const conditionText = `${icon} ${condition.substring(0, 11)}`;
      
      lines.push(
        `│ ${date.padEnd(8)} │ ${conditionText.padEnd(14)} │ ${tempRange.padEnd(7)} │ ${precip.toString().padEnd(6)}%  │ ${wind.padEnd(10)} │`
      );
    });
    
    lines.push('└──────────┴────────────────┴─────────┴──────────┴────────────┘');
    
    return lines.join('\n');
  }

  renderFooter() {
    return chalk.gray('Data provided by Open-Meteo.com • Free weather API for non-commercial use');
  }

  combineBlocks(blocks) {
    const maxHeight = Math.max(...blocks.map(block => block.length));
    const combined = [];
    
    for (let i = 0; i < maxHeight; i++) {
      let line = '';
      blocks.forEach(block => {
        if (i < block.length) {
          line += block[i];
        } else {
          line += ' '.repeat(33);
        }
        line += '  ';
      });
      combined.push(line);
    }
    
    return combined.join('\n');
  }

  getWindDirection(degrees) {
    const directions = ['N', 'NNE', 'NE', 'ENE', 'E', 'ESE', 'SE', 'SSE', 'S', 'SSW', 'SW', 'WSW', 'W', 'WNW', 'NW', 'NNW'];
    const index = Math.round(degrees / 22.5) % 16;
    return directions[index];
  }

  renderSimple(locationData, currentWeather, units = 'imperial', hideFooter = false) {
    const temp = Math.round(currentWeather.temperature);
    const description = weatherService.getWeatherDescription(currentWeather.weatherCode);
    const icon = weatherService.getWeatherIcon(currentWeather.weatherCode, currentWeather.isDay);
    const tempUnit = unitConverter.getTemperatureUnit(units);
    const speedUnit = unitConverter.getSpeedUnit(units);
    const precipUnit = unitConverter.getPrecipitationUnit(units);
    
    const windSpeed = Math.round(currentWeather.windSpeed);
    const precipitation = units === 'imperial' ? currentWeather.precipitation.toFixed(2) : currentWeather.precipitation.toFixed(1);
    
    const lines = [
      `Weather in ${locationData.name}, ${locationData.country}`,
      ``,
      `${icon} ${description}`,
      `🌡️  ${temp}${tempUnit} (feels like ${Math.round(currentWeather.feelsLike)}${tempUnit})`,
      `💨 ${windSpeed} ${speedUnit} ${this.getWindDirection(currentWeather.windDirection)}`,
      `💧 ${currentWeather.humidity}% humidity`,
      `☔ ${precipitation}${precipUnit} precipitation`
    ];
    
    if (!hideFooter) {
      lines.push('');
      lines.push(chalk.gray('Data from Open-Meteo.com'));
    }
    
    // Add two newlines at the end for proper terminal spacing
    return lines.join('\n') + '\n\n';
  }
}

module.exports = new AsciiWeatherRenderer();