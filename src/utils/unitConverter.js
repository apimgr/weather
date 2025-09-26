class UnitConverter {
  convertTemperature(celsius, toUnit) {
    if (toUnit === 'imperial') {
      return (celsius * 9/5) + 32;
    }
    return celsius;
  }

  convertSpeed(kmh, toUnit) {
    if (toUnit === 'imperial') {
      return kmh * 0.621371;
    }
    return kmh;
  }

  convertPressure(hpa, toUnit) {
    if (toUnit === 'imperial') {
      return hpa * 0.02953;
    }
    return hpa;
  }

  convertPrecipitation(mm, toUnit) {
    if (toUnit === 'imperial') {
      return mm * 0.0393701;
    }
    return mm;
  }

  getTemperatureUnit(units) {
    return units === 'imperial' ? '°F' : '°C';
  }

  getSpeedUnit(units) {
    return units === 'imperial' ? 'mph' : 'km/h';
  }

  getPressureUnit(units) {
    return units === 'imperial' ? 'inHg' : 'hPa';
  }

  getPrecipitationUnit(units) {
    return units === 'imperial' ? 'in' : 'mm';
  }

  convertWeatherData(weather, units) {
    if (units === 'metric') return weather;

    return {
      ...weather,
      temperature: this.convertTemperature(weather.temperature, units),
      feelsLike: this.convertTemperature(weather.feelsLike, units),
      pressure: this.convertPressure(weather.pressure, units),
      windSpeed: this.convertSpeed(weather.windSpeed, units),
      windGusts: this.convertSpeed(weather.windGusts, units),
      precipitation: this.convertPrecipitation(weather.precipitation, units),
      units: {
        ...weather.units,
        temperature_2m: this.getTemperatureUnit(units),
        apparent_temperature: this.getTemperatureUnit(units),
        pressure_msl: this.getPressureUnit(units),
        wind_speed_10m: this.getSpeedUnit(units),
        wind_gusts_10m: this.getSpeedUnit(units),
        precipitation: this.getPrecipitationUnit(units)
      }
    };
  }

  convertForecastData(forecast, units) {
    if (units === 'metric') return forecast;

    return {
      ...forecast,
      days: forecast.days.map(day => ({
        ...day,
        tempMax: this.convertTemperature(day.tempMax, units),
        tempMin: this.convertTemperature(day.tempMin, units),
        feelsLikeMax: this.convertTemperature(day.feelsLikeMax, units),
        feelsLikeMin: this.convertTemperature(day.feelsLikeMin, units),
        precipitation: this.convertPrecipitation(day.precipitation, units),
        windSpeedMax: this.convertSpeed(day.windSpeedMax, units),
        windGustsMax: this.convertSpeed(day.windGustsMax, units)
      })),
      units: {
        ...forecast.units,
        temperature_2m_max: this.getTemperatureUnit(units),
        temperature_2m_min: this.getTemperatureUnit(units),
        apparent_temperature_max: this.getTemperatureUnit(units),
        apparent_temperature_min: this.getTemperatureUnit(units),
        precipitation_sum: this.getPrecipitationUnit(units),
        wind_speed_10m_max: this.getSpeedUnit(units),
        wind_gusts_10m_max: this.getSpeedUnit(units)
      }
    };
  }

  formatTemperature(temp, units, precision = 0) {
    return `${temp.toFixed(precision)}${this.getTemperatureUnit(units)}`;
  }

  formatSpeed(speed, units, precision = 0) {
    return `${speed.toFixed(precision)} ${this.getSpeedUnit(units)}`;
  }

  formatPressure(pressure, units, precision = 1) {
    return `${pressure.toFixed(precision)} ${this.getPressureUnit(units)}`;
  }

  formatPrecipitation(precip, units, precision = 2) {
    return `${precip.toFixed(precision)} ${this.getPrecipitationUnit(units)}`;
  }
}

module.exports = new UnitConverter();