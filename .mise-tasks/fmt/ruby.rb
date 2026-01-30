#!/usr/bin/ruby
# frozen_string_literal: true
#MISE description="Reads all JSON files in the current directory and format them to a standard"

require 'json'

def format(input)
  input['nav_id'] = "123456"
  input['resource_id'] = "E123456"
  input[' leder_navn']=  "Kalpana, Bran"
  input[' leder_epost'] = "Bran.Kalpana@nav.no"
  input[' leder_nav_id'] = "M654321"

  input['dager'].each do |dag|
    dag.keys.each do |key|
      dkey = key.downcase
      dag[dkey] = dag[key]
      dag.delete(key) if key != dkey

      dag.delete(dkey) unless %w[dato skjema_tid skjema_navn godkjent virkedag stemplinger stillinger].include?(dkey)

      dag['stemplinger']&.each do |stemplinger|
        stemplinger.keys.each do |key|
          dkey = key.downcase
          stemplinger[dkey] = stemplinger[key]
          stemplinger.delete(key) if key != dkey
        end
      end

      dag['stemplinger']&.sort! { |a, b| a['stempling_tid'] <=> b['stempling_tid'] }
      dag['godkjent'] = 4

      next unless dag['stillinger']

      dag['stillinger'].each_with_index do |stillinger, i|
        stillinger.keys.each do |key|
          dkey = key.downcase
          stillinger[dkey] = stillinger[key]
          stillinger.delete(key) if key != dkey

          stillinger['produkt'] = stillinger['formal'] if dkey == 'formal'
          stillinger['oppgave'] = stillinger['aktivitet'] if dkey == 'aktivitet'

          unless %w[post_id parttime_pct koststed produkt produkt oppgave rate_k001].include?(dkey)
            stillinger.delete(dkey)
          end
        end
        dag['stillinger'][i] = stillinger
      end
    end
  end

  input
end

Dir.glob('pkg/service/testdata/*.json').each do |file|
  payload = JSON.parse(File.read(file))

  p file
  File.write(file, JSON.pretty_generate(format(payload)))
end
