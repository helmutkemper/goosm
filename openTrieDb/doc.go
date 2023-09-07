package opentriedb

// key:
// A chave é a estrutura de dado usada para identificar o dado a ser recuperado e é um array de bytes

// pieces:
// A chave é um array de bytes e a piece é formada por um key.split(size).
// A chave foi definida em 4 bytes para facilitar o uso de texto em UTF8/16 e pelo fato da navegação em arquivo por byte
// ser mais lenta

// length
// Quantidade de slots arquivados na linha

// next address
// Endereço da próxima linha a ser usada quando todos os slots estão ocupados

// slot:
// Conjunto de dados formado por piece, address line e flag
// piece: piece salva para a busca
// address line: endereço da próxima linha para continuar a corrida

// linha:
// Uma linha é um array de bytes onde o endereço 0x00 arquiva a primeira piece do primeiro dado salvo no banco e é
// formada por: formado por length (4 bytes) + next address (8 bytes) + kDataLineLength * slot [piece (4 bytes) +
// address line (8 bytes) + flag (1 byte)]

// | pieces                     | explicação                                                                           |
// |----------------------------|--------------------------------------------------------------------------------------|
// | piece terminal             | a piece termina aqui e a busca deve ser encerrada                                    |
// | piece de corrida           | não há piece terminal nesse ponto                                                    |
// | piece mista                | há uma piece terminal, mas, há pieces para continuar a busca                         |
// | piece terminal atrasada    | esta piece é uma cópia da piece anterior e mesma contém o endereço da próxima tabela |

// | flag | value               | explicação                                                                           |
// |------|---------------------|--------------------------------------------------------------------------------------|
// |   00 | continue            | chave de corrida, o endereço aponta para a próxima linha da mesma tabela             |
// |   01 | complete            | chave terminal, o endereço aponta para a próxima tabela                              |
// |   02 | continue / complete | chave mista, o endereço aponta para a próxima linha da mesma tabela                  |
// |   03 | it self complete    | chave terminal contida na linha anterior                                             |

//address, length, next address, [piece, address line, flag], [piece, address line, flag], [piece, address line, flag]
//            int,        int64, [4byte,        int64, byte], [4byte,        int64, byte], [4byte,        int64, byte]

// start()
//address, length, next address, [piece, address line, flag], [piece, address line, flag], [piece, address line, flag]
//     00,     00,           00, [   00,           00,   00], [   00,           00,   00], [   00,           00,   00]
//
// O banco foi criado do zero;
// Um arquivo de dados foi criado;
// Uma linha foi inserida no endereço headerLength;
// headerLength é o endereço de onde todas as buscas começam;
// A linha contém todos os bytes 0x00.

// search("aa")
// define address = headerLength
// readLength() retorna zero, fim da busca

// insert("aa")
//address, length, next address, [piece, address line,  flag], [piece, address line, flag], [piece, address line, flag]
//     00,      1,           00, [    a,           01,    00], [                         ], [                         ]
//     01,      1,           00, [    a,           xx,    01], [                         ], [                         ]
//
// lê length
// incrementa length + 1
// reservar nova linha (address 01)
// procurar slot em branco (retorno 0)
// gravar slot piece, nova linha (address 01), flag 00
//
//

// insert("ab")
//address, length, next address, [piece, address line,  flag], [piece, address line, flag], [piece, address line, flag]
//     00,      1,           00, [    a,           01,    00], [                         ], [                         ]
//     01,      2,           00, [    a,           xx,    01], [    b,           xx,   01], [                         ]

// insert("ac")
//address, length, next address, [piece, address line,  flag], [piece, address line, flag], [piece, address line, flag]
//     00,      1,           00, [    a,           01,   00], [                          ], [                         ]
//     01,      3,           00, [    a,           xx,   01], [     b,           xx,   01], [    c,           xx,   01]

// insert("ad")
//address, length, next address, [piece, address line,  flag], [piece, address line, flag], [piece, address line, flag]
//     00,      1,           00, [    a,           01,    00], [                         ], [                         ]
//     01,      3,           02, [    a,           xx,    01], [    b,           xx,   01], [    c,           xx,   01]
//     02,      1,           00, [    d,           -1,    01], [                         ], [                         ]

// insert("aba")
//address, length, next address, [piece, address line,  flag], [piece, address line, flag], [piece, address line, flag]
//     00,      1,           00, [    a,           01,    00], [                         ], [                         ]
//     01,      3,           02, [    a,           xx,    01], [    b,           03,   02], [    c,           xx,   01]
//     02,      1,           00, [    d,           -1,    01], [                         ], [                         ]
//     03,      2,           00, [    b,           xx,    03], [    a,           xx,   01], [                         ]
//
// arquivar: address 01, slot 02
// address 01 contém length 3 (máximo)
// reservar nova linha (address 03)
// address 01, nextAddress recebe 03
// address 01, slot 02 contém flag 01, trocar para 02
// address 03, slot 0 recebe dados arquivados e flag 03
// address 03, slot 1 recebe piece 'a' e endereço da próxima tabela

// insert("ba")
//address, length, next address, [piece, address line,  flag], [piece, address line, flag], [piece, address line, flag]
//     00,      2,           00, [    a,           01,    00], [    b,           04    00], [                         ]
//     01,      3,           02, [    a,           xx,    01], [    b,           03,   02], [    c,           xx,   01]
//     02,      1,           00, [    d,           -1,    01], [                         ], [                         ]
//     03,      2,           00, [    b,           xx,    03], [    a,           xx,   01], [                         ]
//     04,      1,           00, [    a,           xx,    01], [                         ], [                         ]

// insert("baa")
//address, length, next address, [piece, address line,  flag], [piece, address line, flag], [piece, address line, flag]
//     00,      2,           00, [    a,           01,    00], [    b,           04    00], [                         ]
//     01,      3,           02, [    a,           xx,    01], [    b,           03,   02], [    c,           xx,   01]
//     02,      1,           00, [    d,           -1,    01], [                         ], [                         ]
//     03,      2,           00, [    b,           xx,    03], [    a,           xx,   01], [                         ]
//     04,      1,           00, [    a,           05,    02], [                         ], [                         ]
//     05,      1,           00, [    a,           xx,    03], [    a,           xx,   01], [                         ]
