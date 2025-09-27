const chalk = require('chalk');
const weatherService = require('../services/weatherService');
const unitConverter = require('./unitConverter');

// Force colors for consistent output
chalk.level = 3;

class OneLineRenderer {
  renderOneLine(location, currentWeather, units) {
    const temp = Math.round(currentWeather.temperature);
    const description = weatherService.getWeatherDescription(currentWeather.weatherCode);
    const icon = weatherService.getWeatherIcon(currentWeather.weatherCode, currentWeather.isDay);
    const tempUnit = unitConverter.getTemperatureUnit(units);
    const windSpeed = Math.round(currentWeather.windSpeed);
    const speedUnit = unitConverter.getSpeedUnit(units);
    const windDir = this.getWindDirection(currentWeather.windDirection);
    
    // One-line format for status bars: location temp condition wind
    const oneLine = `${chalk.hex('#8be9fd')(location.name)}: ${chalk.hex('#f1fa8c')(temp + tempUnit)} ${chalk.hex('#50fa7b')(icon)} ${chalk.hex('#bd93f9')(description)} ${chalk.hex('#ff79c6')(windSpeed + speedUnit + ' ' + windDir)} ${chalk.hex('#ffb86c')(currentWeather.humidity + '%')}`;
    
    return oneLine + '\n\n';
  }

  // Exact wttr.in format implementations
  renderFormat1(currentWeather, units) {
    // Format 1: Just weather icon + temp: "🌦  +11⁰C"
    const temp = Math.round(currentWeather.temperature);
    const icon = weatherService.getWeatherIcon(currentWeather.weatherCode, currentWeather.isDay);
    const tempUnit = unitConverter.getTemperatureUnit(units).replace('°', '');
    
    return `${icon}  ${chalk.hex('#f1fa8c')('+' + temp + '⁰' + tempUnit)}\n`;
  }

  renderFormat2(currentWeather, units) {
    // Format 2: Weather + details: "🌦  🌡️ +11°C  🌬️ ↓ 4km/h"
    const temp = Math.round(currentWeather.temperature);
    const icon = weatherService.getWeatherIcon(currentWeather.weatherCode, currentWeather.isDay);
    const tempUnit = unitConverter.getTemperatureUnit(units);
    const windSpeed = Math.round(currentWeather.windSpeed);
    const speedUnit = unitConverter.getSpeedUnit(units);
    
    return `${icon}  🌡️ ${chalk.hex('#f1fa8c')('+' + temp + tempUnit)}  🌬️ ${chalk.hex('#50fa7b')(this.getWindArrow(currentWeather.windDirection) + ' ' + windSpeed + speedUnit)}\n`;
  }

  renderFormat3(location, currentWeather, units) {
    // Format 3: Location + weather: "London, GB:  🌦  +11⁰C"
    const temp = Math.round(currentWeather.temperature);
    const icon = weatherService.getWeatherIcon(currentWeather.weatherCode, currentWeather.isDay);
    const tempUnit = unitConverter.getTemperatureUnit(units).replace('°', '');
    
    return `${chalk.hex('#8be9fd')(this.getShortLocationName(location))}: ${icon}  ${chalk.hex('#f1fa8c')('+' + temp + '⁰' + tempUnit)}\n`;
  }

  renderFormat4(location, currentWeather, units) {
    // Format 4: Location + weather + details: "London, GB:  🌦  🌡️+11°C  🌬️↓4km/h"
    const temp = Math.round(currentWeather.temperature);
    const icon = weatherService.getWeatherIcon(currentWeather.weatherCode, currentWeather.isDay);
    const tempUnit = unitConverter.getTemperatureUnit(units);
    const windSpeed = Math.round(currentWeather.windSpeed);
    const speedUnit = unitConverter.getSpeedUnit(units);
    
    return `${chalk.hex('#8be9fd')(this.getShortLocationName(location))}: ${icon}  🌡️ ${chalk.hex('#f1fa8c')('+' + temp + tempUnit)}  🌬️ ${chalk.hex('#50fa7b')(this.getWindArrow(currentWeather.windDirection) + ' ' + windSpeed + speedUnit)}\n`;
  }

  getWindDirection(degrees) {
    const directions = ['N', 'NNE', 'NE', 'ENE', 'E', 'ESE', 'SE', 'SSE', 'S', 'SSW', 'SW', 'WSW', 'W', 'WNW', 'NW', 'NNW'];
    const index = Math.round(degrees / 22.5) % 16;
    return directions[index];
  }

  getWindArrow(degrees) {
    const arrows = ['↓', '↙', '←', '↖', '↑', '↗', '→', '↘'];
    const index = Math.round(degrees / 45) % 8;
    return arrows[index];
  }

  getShortLocationName(location) {
    // Use the pre-built short name from enhanced location data
    if (location.shortName) {
      return location.shortName;
    }
    
    // Fallback: build short name using available data
    const cityName = location.name;
    const countryCode = location.countryCode || 'XX';
    
    return `${cityName}, ${countryCode}`;
  }

  getFullLocationName(location) {
    // Use the pre-built full name from enhanced location data
    if (location.fullName) {
      return location.fullName;
    }
    
    // Fallback: basic format
    return `${location.name}, ${location.country}`;
  }
}

module.exports = new OneLineRenderer();