#!/usr/bin/env python3
"""
Process Welsh-English TSV file to expand plural forms.
- No space before (): append content (e.g., enw(au) → enw + enwau)
- Space before (): replace word (e.g., car (ceir) → car + ceir)
"""

import re
import sys

def process_plural(welsh_word):
    """
    Process a Welsh word that may contain a plural form in parentheses.
    Returns a list of (welsh, has_plural) tuples.
    """
    # Pattern 1: Space before parentheses - replace
    # Example: car (ceir) -> car, ceir
    match_replace = re.match(r'^(.+?)\s+\(([^)]+)\)$', welsh_word)

    # Pattern 2: No space before parentheses - append (but check content)
    # Example: uned(au) -> uned, unedau
    match_append = re.match(r'^(.+?)\(([^)]+)\)$', welsh_word)

    if match_replace:
        # Replace pattern (space before parens)
        base_word = match_replace.group(1).strip()
        plural_word = match_replace.group(2).strip()
        return [(base_word, False), (plural_word, True)]
    elif match_append:
        # Append pattern (no space before parens)
        base_word = match_append.group(1).strip()
        content = match_append.group(2).strip()

        # Check if content has leading space - indicates replacement
        if match_append.group(2).startswith(' '):
            # Malformed replace pattern like: rhaglen( findni)
            plural_word = content
            return [(base_word, False), (plural_word, True)]
        else:
            # Normal append pattern
            plural_word = base_word + content
            return [(base_word, False), (plural_word, True)]

    # No plural form found
    return [(welsh_word, False)]

def process_english(english_text, is_plural):
    """
    Process English translation to handle plural forms.
    Removes (s), (es), etc. for singular, applies them for plural.
    """
    # Pattern to find plural markers like (s), (es), (ies), etc.
    pattern = r'\(([^)]+)\)'

    if not is_plural:
        # For singular, remove all plural markers
        result = re.sub(pattern, '', english_text)
        # Clean up any double spaces
        result = re.sub(r'\s+', ' ', result).strip()
        return result
    else:
        # For plural, apply the plural markers
        def replace_plural(match):
            marker = match.group(1)
            # Get the word before the marker
            start = match.start()
            before = english_text[:start].rstrip()

            if marker == 's':
                return 's'
            elif marker == 'es':
                return 'es'
            elif marker == 'ies':
                # Replace 'y' with 'ies' if the word ends in 'y'
                return 'ies'
            elif marker.startswith('('):
                # Complete replacement like (children)
                return marker[1:-1]  # Remove outer parens
            else:
                return marker

        result = re.sub(pattern, replace_plural, english_text)
        # Clean up any double spaces
        result = re.sub(r'\s+', ' ', result).strip()
        return result

def process_tsv(input_file, output_file):
    """Process the TSV file and expand plural forms."""
    output_lines = []

    with open(input_file, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    # Process header
    if lines:
        output_lines.append(lines[0].strip())

    # Process data lines
    for line_num, line in enumerate(lines[1:], start=2):
        line = line.strip()
        if not line:
            continue

        parts = line.split('\t')
        if len(parts) < 2:
            # Malformed line, keep as is
            output_lines.append(line)
            continue

        welsh_col = parts[0]
        english_col = '\t'.join(parts[1:])  # In case there are extra tabs

        # Process the Welsh word
        word_forms = process_plural(welsh_col)

        for welsh_word, is_plural in word_forms:
            # Process English translation based on whether it's singular or plural
            english_processed = process_english(english_col, is_plural)
            output_lines.append(f"{welsh_word}\t{english_processed}")

    # Write output
    with open(output_file, 'w', encoding='utf-8') as f:
        for line in output_lines:
            f.write(line + '\n')

    print(f"Processed {len(lines)} input lines")
    print(f"Generated {len(output_lines)} output lines")
    print(f"Output written to: {output_file}")

if __name__ == '__main__':
    input_file = '/data/private/words.tsv'
    output_file = '/data/private/words_expanded.tsv'

    process_tsv(input_file, output_file)
