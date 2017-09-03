# Basic script to generate go text template strings from container code.
# Compatible with Python 2 and 3. Only relies on stdlib.

if __name__ == "__main__":
    template_start_mark = '//template:'
    input_file_name = 'internal/generic_types/containers.go'
    output_file_name = 'internal/templates/templates.go'
    template_package_name = 'templates'
    generic_types = {'GenericVectorType': 'VectorTypeName',
                     'GenericType': 'TypeName',
                     'GenericMapType': 'MapTypeName',
                     'GenericMapItem': 'MapItemTypeName',
                     'GenericMapKeyType': 'MapKeyTypeName',
                     'GenericMapValueType': 'MapValueTypeName',
                     'genericHash': 'MapKeyHashFunc',
                     'GenericSetType': 'SetTypeName'}

    state = 'searching'
    template_name = ''
    template = ''
    templates = {}

    print("Generating templates")
    with open(input_file_name, 'r') as input_file:
        for line in input_file:
            if line.startswith(template_start_mark):
                if state == 'reading':
                    templates[template_name] = template

                template_name = line[len(template_start_mark):].strip()
                template = ''
                state = 'reading'
            else:
                if state == 'reading':
                    template += line

        if state == 'reading':
            templates[template_name] = template

    with open(output_file_name, 'w') as output_file:
        output_file.write('package {}\n\n'.format(template_package_name))
        output_file.write("// NOTE: This file is auto generated, don't edit manually!\n".format(template_package_name))
        for name, template in sorted(list(templates.items())):
            for generic_type, variable_name in generic_types.items():
                template = template.replace(generic_type, '{{{{.{variable_name}}}}}'.format(variable_name=variable_name))
            output_file.write('const {name} string = `'.format(name=name))
            output_file.write(template)
            output_file.write('`\n')

    print("Done!")
