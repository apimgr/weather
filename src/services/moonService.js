const SunCalc = require('suncalc');
const chalk = require('chalk');

// Force colors to be enabled for consistent Dracula theme
chalk.level = 3;

class MoonService {
  constructor() {
    this.phases = [
      'New Moon',
      'Waxing Crescent', 
      'First Quarter',
      'Waxing Gibbous',
      'Full Moon',
      'Waning Gibbous',
      'Last Quarter',
      'Waning Crescent'
    ];

    this.phaseIcons = [
      '🌑', // New Moon
      '🌒', // Waxing Crescent
      '🌓', // First Quarter
      '🌔', // Waxing Gibbous
      '🌕', // Full Moon
      '🌖', // Waning Gibbous
      '🌗', // Last Quarter
      '🌘'  // Waning Crescent
    ];
  }

  getMoonPhase(date = new Date(), latitude = 40.7128, longitude = -74.0060) {
    const moonIllumination = SunCalc.getMoonIllumination(date);
    const moonTimes = SunCalc.getMoonTimes(date, latitude, longitude);
    const moonPosition = SunCalc.getMoonPosition(date, latitude, longitude);
    
    // Calculate phase based on illumination and angle
    const illumination = moonIllumination.fraction;
    const phase = moonIllumination.phase;
    
    // Determine phase index (0-7)
    let phaseIndex;
    if (phase < 0.125) phaseIndex = 0; // New Moon
    else if (phase < 0.25) phaseIndex = 1; // Waxing Crescent
    else if (phase < 0.375) phaseIndex = 2; // First Quarter
    else if (phase < 0.5) phaseIndex = 3; // Waxing Gibbous
    else if (phase < 0.625) phaseIndex = 4; // Full Moon
    else if (phase < 0.75) phaseIndex = 5; // Waning Gibbous
    else if (phase < 0.875) phaseIndex = 6; // Last Quarter
    else phaseIndex = 7; // Waning Crescent

    return {
      phase: this.phases[phaseIndex],
      icon: this.phaseIcons[phaseIndex],
      illumination: Math.round(illumination * 100),
      age: Math.round(phase * 29.53), // Moon cycle is ~29.53 days
      rise: moonTimes.rise,
      set: moonTimes.set,
      altitude: moonPosition.altitude,
      azimuth: moonPosition.azimuth,
      distance: moonPosition.distance
    };
  }

  renderMoonReport(location, date = new Date()) {
    const moon = this.getMoonPhase(date, location.latitude, location.longitude);
    const sunTimes = SunCalc.getTimes(date, location.latitude, location.longitude);
    
    const lines = [];
    
    // Header with Dracula colors
    lines.push(chalk.hex('#f1fa8c').bold(`Moon phase report: ${location.name.toLowerCase()}`));
    lines.push('');
    
    // Current moon phase with enhanced ASCII art
    lines.push(...this.renderMoonAsciiArt(moon, sunTimes));
    lines.push('');
    lines.push('');
    
    // Moon information table (weather-style)
    lines.push(...this.renderMoonInfoTable(moon, sunTimes));
    lines.push('');
    
    // Footer
    lines.push(chalk.hex('#6272a4')('Moon and sun data calculated using SunCalc • Console Weather Service'));
    
    // Add two newlines at the end for proper terminal spacing
    return lines.join('\n') + '\n\n';
  }

  renderMoonAsciiArt(moon, sunTimes) {
    const ascii = this.getDetailedMoonArt(moon.illumination);
    const phaseColor = this.getMoonColor(moon.illumination);
    
    // Current moon phase with ASCII art (weather-style layout)
    return [
      `${chalk.hex(phaseColor)(ascii.line1)}     ${chalk.hex('#8be9fd')(moon.phase)}`,
      `${chalk.hex(phaseColor)(ascii.line2)}     ${chalk.hex('#f1fa8c')(`${moon.illumination}% illuminated`)}`,
      `${chalk.hex(phaseColor)(ascii.line3)}     ${chalk.hex('#bd93f9')(`Age: ${moon.age} days`)}`,
      `${chalk.hex(phaseColor)(ascii.line4)}     ${chalk.hex('#50fa7b')(`${moon.icon} Lunar cycle`)}`,
      `${chalk.hex(phaseColor)(ascii.line5)}     ${chalk.hex('#ff79c6')(`Next: ${this.getNextPhase(moon.age)}`)}`
    ];
  }

  renderMoonInfoTable(moon, sunTimes) {
    const lines = [];
    
    // Create a weather-style information table
    const moonBlock = [
      chalk.hex('#bd93f9')('┌─────────────────────────────┐'),
      chalk.hex('#bd93f9')('│') + chalk.hex('#ffb86c').bold('          Moon Data          ') + chalk.hex('#bd93f9')('│'),
      chalk.hex('#bd93f9')('│') + `  Phase: ${chalk.hex('#f1fa8c')(moon.phase.padEnd(16))}` + chalk.hex('#bd93f9')('│'),
      chalk.hex('#bd93f9')('│') + `  Illumination: ${chalk.hex('#8be9fd')((moon.illumination + '%').padEnd(11))}` + chalk.hex('#bd93f9')('│'),
      chalk.hex('#bd93f9')('│') + `  Age: ${chalk.hex('#50fa7b')((moon.age + ' days').padEnd(18))}` + chalk.hex('#bd93f9')('│'),
      moon.rise && moon.set ? 
        chalk.hex('#bd93f9')('│') + `  Rise: ${chalk.hex('#ff79c6')(moon.rise.toLocaleTimeString().padEnd(17))}` + chalk.hex('#bd93f9')('│') :
        chalk.hex('#bd93f9')('│') + `  Rise: ${chalk.hex('#6272a4')('Not available'.padEnd(17))}` + chalk.hex('#bd93f9')('│'),
      moon.rise && moon.set ?
        chalk.hex('#bd93f9')('│') + `  Set: ${chalk.hex('#ff79c6')(moon.set.toLocaleTimeString().padEnd(18))}` + chalk.hex('#bd93f9')('│') :
        chalk.hex('#bd93f9')('│') + `  Set: ${chalk.hex('#6272a4')('Not available'.padEnd(18))}` + chalk.hex('#bd93f9')('│'),
      chalk.hex('#bd93f9')('└─────────────────────────────┘')
    ];

    const sunBlock = [
      chalk.hex('#bd93f9')('┌─────────────────────────────┐'),
      chalk.hex('#bd93f9')('│') + chalk.hex('#ffb86c').bold('          Sun Data           ') + chalk.hex('#bd93f9')('│'),
      chalk.hex('#bd93f9')('│') + `  Sunrise: ${chalk.hex('#f1fa8c')(sunTimes.sunrise.toLocaleTimeString().padEnd(14))}` + chalk.hex('#bd93f9')('│'),
      chalk.hex('#bd93f9')('│') + `  Sunset: ${chalk.hex('#ff5555')(sunTimes.sunset.toLocaleTimeString().padEnd(15))}` + chalk.hex('#bd93f9')('│'),
      chalk.hex('#bd93f9')('│') + `  Solar Noon: ${chalk.hex('#50fa7b')(sunTimes.solarNoon.toLocaleTimeString().padEnd(11))}` + chalk.hex('#bd93f9')('│'),
      chalk.hex('#bd93f9')('│') + `  Day Length: ${chalk.hex('#8be9fd')(this.getDayLength(sunTimes).padEnd(11))}` + chalk.hex('#bd93f9')('│'),
      chalk.hex('#bd93f9')('│') + `  Season: ${chalk.hex('#bd93f9')(this.getSeason(new Date()).padEnd(15))}` + chalk.hex('#bd93f9')('│'),
      chalk.hex('#bd93f9')('└─────────────────────────────┘')
    ];

    return this.combineBlocks([moonBlock, sunBlock]);
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
    
    return combined;
  }

  getMoonAsciiArt(illumination) {
    if (illumination < 10) {
      // New Moon
      return {
        line1: '     ',
        line2: '  ●  ',
        line3: '     ',
        line4: '     ',
        line5: '     '
      };
    } else if (illumination < 30) {
      // Crescent
      return {
        line1: '     ',
        line2: ' ◐   ',
        line3: '     ',
        line4: '     ',
        line5: '     '
      };
    } else if (illumination < 60) {
      // Quarter
      return {
        line1: '     ',
        line2: ' ◑   ',
        line3: '     ',
        line4: '     ',
        line5: '     '
      };
    } else if (illumination < 90) {
      // Gibbous
      return {
        line1: '     ',
        line2: ' ◒   ',
        line3: '     ',
        line4: '     ',
        line5: '     '
      };
    } else {
      // Full Moon
      return {
        line1: '     ',
        line2: ' ○   ',
        line3: '     ',
        line4: '     ',
        line5: '     '
      };
    }
  }

  getDetailedMoonArt(illumination) {
    if (illumination < 5) {
      // New Moon
      return {
        line1: '             ',
        line2: '     ●●●     ',
        line3: '   ●●   ●●   ',
        line4: '   ●     ●   ',
        line5: '   ●●   ●●   '
      };
    } else if (illumination < 25) {
      // Waxing Crescent
      return {
        line1: '             ',
        line2: '     ◐◐◐     ',
        line3: '   ◐◐   ●●   ',
        line4: '   ◐     ●   ',
        line5: '   ◐◐   ●●   '
      };
    } else if (illumination < 45) {
      // First Quarter
      return {
        line1: '             ',
        line2: '     ◑◑◑     ',
        line3: '   ◑◑   ●●   ',
        line4: '   ◑     ●   ',
        line5: '   ◑◑   ●●   '
      };
    } else if (illumination < 75) {
      // Waxing Gibbous
      return {
        line1: '             ',
        line2: '     ○○◒     ',
        line3: '   ○○   ◒◒   ',
        line4: '   ○     ◒   ',
        line5: '   ○○   ◒◒   '
      };
    } else if (illumination < 95) {
      // Full Moon
      return {
        line1: '             ',
        line2: '     ○○○     ',
        line3: '   ○○   ○○   ',
        line4: '   ○     ○   ',
        line5: '   ○○   ○○   '
      };
    } else {
      // Waning phases (mirror)
      return {
        line1: '             ',
        line2: '     ○○○     ',
        line3: '   ○○   ○○   ',
        line4: '   ○     ○   ',
        line5: '   ○○   ○○   '
      };
    }
  }

  getMoonColor(illumination) {
    if (illumination < 10) return '#6272a4'; // Dark gray for new moon
    if (illumination < 50) return '#f8f8f2'; // Light gray for crescents
    if (illumination < 90) return '#f1fa8c'; // Yellow for gibbous
    return '#8be9fd'; // Cyan for full moon
  }

  getNextPhase(age) {
    if (age < 7) return 'First Quarter';
    if (age < 14) return 'Full Moon';
    if (age < 21) return 'Last Quarter'; 
    return 'New Moon';
  }

  getDayLength(sunTimes) {
    const length = sunTimes.sunset - sunTimes.sunrise;
    const hours = Math.floor(length / (1000 * 60 * 60));
    const minutes = Math.floor((length % (1000 * 60 * 60)) / (1000 * 60));
    return `${hours}h ${minutes}m`;
  }

  getSeason(date) {
    const month = date.getMonth();
    const day = date.getDate();
    
    if ((month === 11 && day >= 21) || month === 0 || month === 1 || (month === 2 && day < 20)) return 'Winter';
    if ((month === 2 && day >= 20) || month === 3 || month === 4 || (month === 5 && day < 21)) return 'Spring';
    if ((month === 5 && day >= 21) || month === 6 || month === 7 || (month === 8 && day < 22)) return 'Summer';
    return 'Autumn';
  }
}

module.exports = new MoonService();