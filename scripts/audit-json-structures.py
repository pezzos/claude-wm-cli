#!/usr/bin/env python3
"""
Script d'audit pour vérifier la cohérence entre les structures Go et les schémas JSON
"""

import os
import json
import re
import subprocess
from pathlib import Path
from typing import Dict, List, Tuple

# Mapping des fichiers JSON vers leurs schémas
JSON_TO_SCHEMA = {
    "epics.json": "epics.schema.json",
    "stories.json": "stories.schema.json", 
    "current-story.json": "current-story.schema.json",
    "current-epic.json": "current-epic.schema.json",
    "current-task.json": "current-task.schema.json",
    "iterations.json": "iterations.schema.json",
    "metrics.json": "metrics.schema.json"
}

def find_go_files_with_json_parsing():
    """Trouve tous les fichiers Go qui parsent du JSON"""
    result = subprocess.run(['find', '.', '-name', '*.go', '-exec', 'grep', '-l', 'json.Unmarshal', '{}', ';'], 
                          capture_output=True, text=True)
    return result.stdout.strip().split('\n') if result.stdout.strip() else []

def extract_struct_definitions(go_file: str) -> List[Dict]:
    """Extrait les définitions de structures Go d'un fichier"""
    try:
        with open(go_file, 'r') as f:
            content = f.read()
    except:
        return []
    
    # Cherche les structures avec des tags json
    struct_pattern = r'type\s+(\w+)\s+struct\s*\{([^}]+)\}'
    json_tag_pattern = r'(\w+)\s+([^`\n]+)\s+`json:"([^"]+)"`'
    
    structs = []
    for struct_match in re.finditer(struct_pattern, content, re.DOTALL):
        struct_name = struct_match.group(1)
        struct_body = struct_match.group(2)
        
        fields = []
        for field_match in re.finditer(json_tag_pattern, struct_body):
            field_name = field_match.group(1)
            field_type = field_match.group(2).strip()
            json_tag = field_match.group(3)
            
            fields.append({
                'name': field_name,
                'type': field_type,
                'json_tag': json_tag
            })
        
        if fields:  # Seulement les structs avec des tags JSON
            structs.append({
                'name': struct_name,
                'file': go_file,
                'fields': fields
            })
    
    return structs

def analyze_json_usage_in_file(go_file: str) -> List[Dict]:
    """Analyse l'utilisation de JSON dans un fichier Go"""
    try:
        with open(go_file, 'r') as f:
            content = f.read()
    except:
        return []
    
    usages = []
    
    # Cherche les utilisations de nos fichiers JSON spécifiques
    for json_file in JSON_TO_SCHEMA.keys():
        if json_file in content:
            # Cherche les structures inline qui parsent ce JSON
            lines = content.split('\n')
            for i, line in enumerate(lines):
                if json_file in line and 'json.Unmarshal' in content[max(0, content.find(line)-500):content.find(line)+500]:
                    usages.append({
                        'json_file': json_file,
                        'go_file': go_file,
                        'line_context': line.strip(),
                        'line_number': i + 1
                    })
    
    return usages

def load_schema(schema_file: str) -> Dict:
    """Charge un schéma JSON"""
    schema_path = f"internal/config/system/commands/templates/schemas/{schema_file}"
    try:
        with open(schema_path, 'r') as f:
            return json.load(f)
    except:
        return {}

def analyze_schema_structure(schema: Dict, path: str = "") -> Dict:
    """Analyse la structure d'un schéma JSON récursivement"""
    structure = {}
    
    if 'properties' in schema:
        for prop_name, prop_def in schema['properties'].items():
            current_path = f"{path}.{prop_name}" if path else prop_name
            
            if 'type' in prop_def:
                if prop_def['type'] == 'object':
                    if 'patternProperties' in prop_def:
                        # C'est un objet avec des clés dynamiques (map)
                        structure[prop_name] = {
                            'type': 'map',
                            'key_pattern': list(prop_def['patternProperties'].keys())[0] if prop_def['patternProperties'] else None,
                            'value_type': analyze_schema_structure(list(prop_def['patternProperties'].values())[0] if prop_def['patternProperties'] else {})
                        }
                    elif 'properties' in prop_def:
                        # C'est un objet avec des propriétés fixes
                        structure[prop_name] = {
                            'type': 'object',
                            'properties': analyze_schema_structure(prop_def)
                        }
                    else:
                        structure[prop_name] = {'type': 'object'}
                elif prop_def['type'] == 'array':
                    # C'est un tableau
                    items_def = prop_def.get('items', {})
                    structure[prop_name] = {
                        'type': 'array',
                        'items': analyze_schema_structure(items_def) if 'type' in items_def else items_def
                    }
                else:
                    structure[prop_name] = {'type': prop_def['type']}
    
    return structure

def detect_inconsistencies():
    """Détecte les incohérences entre Go structs et schémas JSON"""
    print("🔍 AUDIT DES STRUCTURES JSON/GO")
    print("=" * 50)
    
    go_files = find_go_files_with_json_parsing()
    print(f"📁 Fichiers Go analysés: {len(go_files)}")
    
    inconsistencies = []
    
    # Analyse chaque fichier JSON et son schéma
    for json_file, schema_file in JSON_TO_SCHEMA.items():
        print(f"\n📋 Analyse de {json_file}")
        print("-" * 30)
        
        # Charge le schéma
        schema = load_schema(schema_file)
        if not schema:
            print(f"❌ Impossible de charger le schéma {schema_file}")
            continue
            
        # Analyse la structure du schéma
        schema_structure = analyze_schema_structure(schema)
        print(f"📐 Structure du schéma: {list(schema_structure.keys())}")
        
        # Trouve les fichiers Go qui utilisent ce JSON
        go_usages = []
        for go_file in go_files:
            usages = analyze_json_usage_in_file(go_file)
            for usage in usages:
                if usage['json_file'] == json_file:
                    go_usages.append(usage)
        
        print(f"🔧 Fichiers Go qui parsent ce JSON: {len(go_usages)}")
        for usage in go_usages:
            print(f"   - {usage['go_file']}:{usage['line_number']}")
            
            # Cas spécifique connu: stories.json avec map vs array
            if json_file == "stories.json":
                with open(usage['go_file'], 'r') as f:
                    go_content = f.read()
                    
                if 'Stories []struct' in go_content:
                    inconsistencies.append({
                        'json_file': json_file,
                        'go_file': usage['go_file'],
                        'issue': 'Stories défini comme []struct mais le schéma définit un objet (map)',
                        'fix': 'Changer []struct en map[string]struct'
                    })
                    print(f"   ❌ PROBLÈME: Stories défini comme []struct au lieu de map[string]struct")
                elif 'Stories map[string]struct' in go_content:
                    print(f"   ✅ OK: Stories correctement défini comme map[string]struct")
    
    # Rapport final
    print(f"\n📊 RÉSUMÉ DE L'AUDIT")
    print("=" * 50)
    if inconsistencies:
        print(f"❌ {len(inconsistencies)} incohérence(s) détectée(s):")
        for i, issue in enumerate(inconsistencies, 1):
            print(f"\n{i}. {issue['json_file']} dans {issue['go_file']}")
            print(f"   Problème: {issue['issue']}")
            print(f"   Solution: {issue['fix']}")
    else:
        print("✅ Aucune incohérence détectée!")
    
    return inconsistencies

if __name__ == "__main__":
    detect_inconsistencies()