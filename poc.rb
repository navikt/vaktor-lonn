# frozen_string_literal: true

require 'date'
require 'json'
require 'pp'
require 'time'

# Beredskap på mandager-fredager er maksimalt 17 timer per døgn i sommertid og maksimalt 16 timer og 15 minutter i vintertid, men 16 timer og 30 minutter per døgn i fast arbeidstid året rundt."
# Ref avtalen kap. 5.
# Se også kapittel 9, f.eks. om beredskapstillegget:
# "Beredskapstillegget betales også for den tiden som det betales overtid for grunnet en beredskapshendelse, jf. HTA § 17

def get_periode
  {
    "ident": 'E152435',
    "satser": {
      "lørsøn": 55,
      "0620": 10,
      "2006": 20,
      "utvidet": 15
    },
    "periodes": {
      "14.03.2022": {
        "fra": '0000',
        "til": '2400',
        "helligdag": false
      },
      "15.03.2022": {
        "fra": '0000',
        "til": '2400',
        "helligdag": false
      },
      "16.03.2022": {
        "fra": '0000',
        "til": '2400',
        "helligdag": false
      },
      "17.03.2022": {
        "fra": '0000',
        "til": '2400',
        "helligdag": true
      },
      "18.03.2022": {
        "fra": '0000',
        "til": '2400',
        "helligdag": false
      },
      "19.03.2022": {
        "fra": '0000',
        "til": '2400',
        "helligdag": false
      },
      "20.03.2022": {
        "fra": '0000',
        "til": '2400',
        "helligdag": false
      }
    }
  }
end

def get_minwintid
  {
    "14.03.2022": ['07:15-15:33'],
    "15.03.2022": ['07:15-15:57'],
    "16.03.2022": ['01:00-03:00', '07:31-15:33'],
    "17.03.2022": ['07:55-16:10'],
    "18.03.2022": ['07:30-16:19'],
    "19.03.2022": [],
    "20.03.2022": []
  }
end

def get_agresso
  # koststed, personalleder, årslønn
  ['123321', 'Knut Mork', '500_000']
end

def time_to_minutes(time)
  t = Time.parse(time)
  t.hour * 60 + t.min
end

def ranges_overlap?(rA, rB)
  rB.begin <= rA.end && rA.begin <= rB.end
end

def count_range_overlap(rA, rB)
  return rB.count if rA.cover?(rB)

  rangeStart = rB.begin < rA.begin ? rA.begin : rB.begin
  rangeEnd = rB.end > rA.end ? rA.end : rB.end
  (rangeStart...rangeEnd).count
end

def parse_periode(periodes, timecard)
  minutes = {}

  # klokkeslett blir gjort om til minutter
  nightTimeRange = (0...360) # 00-06
  dayTimeRange = (360...1200) # 06-20
  eveningTimeRange = (1200...1440) # 20-00
  preWorkRange = (360...420) # 06-07
  postWorkRange = (1020...1200) # 17-20
  weekendRange = (0...1440) # 00-24

  # check for summertime/winterTime
  today = Date.today
  lastDayInMarch = Date.new(today.year, 3, 31)
  summerTime = lastDayInMarch - lastDayInMarch.wday
  lastDayInOctober = Date.new(today.year, 10, 31)
  winterTime = lastDayInOctober - lastDayInOctober.wday
  if today == summerTime
    # add an extra hour
    nightTimeRange = (0...420)
  elsif today == winterTime
    # remove an hour
    nightTimeRange = (0...300)
  end

  periodes.each do |day, periode|
    date = Date.parse(day.to_s)

    minutes[day] = {
      "dayMinutes": dayTimeRange.count,
      "nightMinutes": nightTimeRange.count + eveningTimeRange.count
    }

    if periode[:helligdag] || date.saturday? || date.sunday?
      minutes[day][:weekendMinutes] = weekendRange.count
      minutes[day][:helgetillegg] = true
    else
      minutes[day][:workMinutes] = preWorkRange.count + postWorkRange.count
      minutes[day][:helgetillegg] = false
    end

    workingHours = timecard[day]
    # TODO: støtte forskjellige vaktplaner, denne koden funker kun for døgnvakt
    workingHours.each do |values|
      from, to = values.split('-')
      workRange = (time_to_minutes(from)...time_to_minutes(to))

      if ranges_overlap?(nightTimeRange, workRange)
        minutes[day][:nightMinutes] -= count_range_overlap(nightTimeRange, workRange)
      end

      if ranges_overlap?(eveningTimeRange, workRange)
        minutes[day][:nightMinutes] -= count_range_overlap(eveningTimeRange, workRange)
      end

      if ranges_overlap?(dayTimeRange, workRange)
        minutes[day][:dayMinutes] -= count_range_overlap(dayTimeRange, workRange)
      end

      # ikke i helger eller helligdager!
      if periode[:helligdag] || date.saturday? || date.sunday?
        minutes[day][:weekendMinutes] -= workRange.count
      else
        if ranges_overlap?(preWorkRange, workRange)
          minutes[day][:workMinutes] -= count_range_overlap(preWorkRange, workRange)
        end

        if ranges_overlap?(postWorkRange, workRange)
          minutes[day][:workMinutes] -= count_range_overlap(postWorkRange, workRange)
        end
      end
    end
  end

  pp minutes
end

def calculate_compensation(worked)
  p 'Compensation'
  dayHours = (worked.sum { |_, values| values[:dayMinutes] } / 60.0).round
  nightHours = (worked.sum { |_, values| values[:nightMinutes] } / 60.0).round
  workHours = (worked.sum { |_, values| values[:workMinutes] || 0 } / 60.0).round
  weekendHours = (worked.sum { |_, values| values[:weekendMinutes] || 0 } / 60.0).round
  p dayHours, nightHours, workHours, weekendHours
  p dayCompensation = dayHours * 10.0
  p nightCompensation = nightHours * 20.0
  p workCompensation = (workHours * 15.0) / 5
  p weekendCompensation = (weekendHours * 55.0) / 5

  print 'Total: '
  dayCompensation + nightCompensation + workCompensation + weekendCompensation
end

def calculate_overtime(worked, salary)
  p 'Overtime'
  overtimeWeekendHours = ((worked.sum {|_, values| values[:helgetillegg] ? values[:dayMinutes] : 0} / 60.0) +
                          (worked.sum {|_, values| values[:helgetillegg] ? values[:nightMinutes] : 0} / 60.0)).round
  overtimeDayWorkHours = (worked.sum { |_, values| values[:helgetillegg] ? 0 : values[:dayMinutes] } / 60.0).round
  overtimeNightHours = (worked.sum { |_, values| values[:helgetillegg] ? 0 : values[:nightMinutes] } / 60.0).round
  p overtimeWeekendHours, overtimeDayWorkHours, overtimeNightHours
  # =IF(B6="Vanlig";(AZ6*A$33+(AX6+AY6)*A$31)/5;(AX6+AY6+AZ6)*A$31/5)
  # b6 er vanlig
  # 50% == (lønn/1850)*1,5
  # 100% == (lønn/1850)*2
  p ots50 = ((salary / 1850.0) * 1.5).round(2)
  p ots100 = ((salary / 1850.0) * 2).round(2)
  # arbeidsdag = (AZ6*A$33+(AX6+AY6)*A$31)/5 == ((dayHours*ots50)+(nightHours*ots100)/5
  # helg = (AX6+AY6+AZ6)*A$31/5 == (dayHours+nightHours)*ots100/5
  p workOvertime = (overtimeDayWorkHours * ots50 + overtimeNightHours * ots100) / 5
  p weekendOvertime = overtimeWeekendHours * ots100 / 5

  print 'Total: '
  (workOvertime + weekendOvertime).round(2)
end

def calculate_earnings(worked, salary)
  p compensation = calculate_compensation(worked)
  p overtime = calculate_overtime(worked, salary)

  # endelig sum
  p 'Payout'
  p (compensation + overtime).round(2)
end

def main
  periode = get_periode
  timecard = get_minwintid
  koststed, personalleder, salary = get_agresso
  minutesWorked = parse_periode(periode[:periodes], timecard)
  calculate_earnings(minutesWorked, salary.to_i)
end

main
